// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dtl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	hashiVMSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/virtualmachines"

	hashiLabsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/devtestlab/2018-09-15/labs"
	hashiDTLVMSDK "github.com/hashicorp/go-azure-sdk/resource-manager/devtestlab/2018-09-15/virtualmachines"
	"github.com/hashicorp/go-azure-sdk/resource-manager/network/2022-09-01/networkinterfaces"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/constants"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

type StepDeployTemplate struct {
	client  *AzureClient
	deploy  func(ctx context.Context, resourceGroupName string, deploymentName string, state multistep.StateBag) error
	disk    func(ctx context.Context, subscriptionId string, resourceGroupName string, computeName string) (string, string, error)
	say     func(message string)
	error   func(e error)
	config  *Config
	factory templateFactoryFuncDtl
	name    string
}

func NewStepDeployTemplate(client *AzureClient, ui packersdk.Ui, config *Config, deploymentName string, factory templateFactoryFuncDtl) *StepDeployTemplate {
	var step = &StepDeployTemplate{
		client:  client,
		say:     func(message string) { ui.Say(message) },
		error:   func(e error) { ui.Error(e.Error()) },
		config:  config,
		factory: factory,
		name:    deploymentName,
	}

	step.deploy = step.deployTemplate
	step.disk = step.getImageDetails
	return step
}

func (s *StepDeployTemplate) deployTemplate(ctx context.Context, resourceGroupName string, deploymentName string, state multistep.StateBag) error {

	subscriptionId := s.config.ClientConfig.SubscriptionID
	labName := s.config.LabName

	// TODO Talk to Tom(s) about this, we have to have two different Labs IDs in different calls, so we can probably move this into the commonids package
	labResourceId := hashiDTLVMSDK.NewLabID(subscriptionId, resourceGroupName, labName)
	labId := hashiLabsSDK.NewLabID(subscriptionId, s.config.tmpResourceGroupName, labName)
	vmlistPage, err := s.client.DtlMetaClient.VirtualMachines.List(ctx, labResourceId, hashiDTLVMSDK.DefaultListOperationOptions())
	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	vmList := vmlistPage.Model
	for _, vm := range *vmList {
		if *vm.Name == s.config.tmpComputeName {
			return fmt.Errorf("Error: Virtual Machine %s already exists. Please use another name", s.config.tmpComputeName)
		}
	}

	s.say(fmt.Sprintf("Creating Virtual Machine %s", s.config.tmpComputeName))
	labMachine, err := s.factory(s.config)
	if err != nil {
		return err
	}

	err = s.client.DtlMetaClient.Labs.CreateEnvironmentThenPoll(ctx, labId, *labMachine)
	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	expand := "Properties($expand=ComputeVm,Artifacts,NetworkInterface)"
	vmResourceId := hashiDTLVMSDK.NewVirtualMachineID(subscriptionId, s.config.tmpResourceGroupName, labName, s.config.tmpComputeName)
	vm, err := s.client.DtlMetaClient.VirtualMachines.Get(ctx, vmResourceId, hashiDTLVMSDK.GetOperationOptions{Expand: &expand})
	if err != nil {
		s.say(s.client.LastError.Error())
	}

	// set tmpFQDN to the PrivateIP or to the real FQDN depending on
	// publicIP being allowed or not
	if s.config.DisallowPublicIP {
		interfaceID := commonids.NewNetworkInterfaceID(subscriptionId, resourceGroupName, s.config.tmpNicName)
		resp, err := s.client.NetworkMetaClient.NetworkInterfaces.Get(ctx, interfaceID, networkinterfaces.DefaultGetOperationOptions())
		if err != nil {
			s.say(s.client.LastError.Error())
			return err
		}
		// TODO This operation seems kinda off, but I don't wanna spend time digging into it right now
		s.config.tmpFQDN = *(*resp.Model.Properties.IPConfigurations)[0].Properties.PrivateIPAddress
	} else {
		s.config.tmpFQDN = *vm.Model.Properties.Fqdn
	}
	s.say(fmt.Sprintf(" -> VM FQDN/IP : '%s'", s.config.tmpFQDN))
	state.Put(constants.SSHHost, s.config.tmpFQDN)

	// In a windows VM, add the winrm artifact. Doing it after the machine has been
	// created allows us to use its IP address as FQDN
	if strings.ToLower(s.config.OSType) == "windows" {
		// Add mandatory Artifact
		var winrma = "windows-winrm"
		var artifactid = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DevTestLab/labs/%s/artifactSources/public repo/artifacts/%s",
			s.config.ClientConfig.SubscriptionID,
			s.config.tmpResourceGroupName,
			s.config.LabName,
			winrma)

		var hostname = "hostName"
		dp := &hashiDTLVMSDK.ArtifactParameterProperties{}
		dp.Name = &hostname
		dp.Value = &s.config.tmpFQDN
		dparams := []hashiDTLVMSDK.ArtifactParameterProperties{*dp}

		winrmArtifact := &hashiDTLVMSDK.ArtifactInstallProperties{
			ArtifactTitle: &winrma,
			ArtifactId:    &artifactid,
			Parameters:    &dparams,
		}

		dtlArtifacts := []hashiDTLVMSDK.ArtifactInstallProperties{*winrmArtifact}
		dtlArtifactsRequest := hashiDTLVMSDK.ApplyArtifactsRequest{Artifacts: &dtlArtifacts}

		// TODO this was an infinite loop, I have seen apply artifacts fail
		// But this needs a bit further validation into why it fails and
		// How we can avoid the need for a retry backoff
		// But a retry backoff is much more preferable to an infinite loop

		retryConfig := retry.Config{
			Tries:      5,
			RetryDelay: (&retry.Backoff{InitialBackoff: 5 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 1.5}).Linear,
		}
		err = retryConfig.Run(ctx, func(ctx context.Context) error {
			err := s.client.DtlMetaClient.VirtualMachines.ApplyArtifactsThenPoll(ctx, vmResourceId, dtlArtifactsRequest)
			if err != nil {
				s.say("WinRM artifact deployment failed, retrying")
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	xs := strings.Split(*vm.Model.Properties.ComputeId, "/")
	s.config.VMCreationResourceGroup = xs[4]

	// Resuing the Resource group name from common constants as all steps depend on it.
	state.Put(constants.ArmResourceGroupName, s.config.VMCreationResourceGroup)

	s.say(fmt.Sprintf(" -> VM ResourceGroupName : '%s'", s.config.VMCreationResourceGroup))

	return err
}

func (s *StepDeployTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.say("Deploying deployment template ...")

	var resourceGroupName = state.Get(constants.ArmResourceGroupName).(string)

	s.say(fmt.Sprintf(" -> Lab ResourceGroupName : '%s'", resourceGroupName))

	return processStepResult(
		s.deploy(ctx, resourceGroupName, s.name, state),
		s.error, state)
}

func (s *StepDeployTemplate) getImageDetails(ctx context.Context, subscriptionId string, resourceGroupName string, computeName string) (string, string, error) {
	//We can't depend on constants.ArmOSDiskVhd being set
	var imageName, imageType string
	vmID := hashiVMSDK.NewVirtualMachineID(subscriptionId, resourceGroupName, computeName)
	vm, err := s.client.VirtualMachinesClient.Get(ctx, vmID, hashiVMSDK.DefaultGetOperationOptions())
	if err != nil {
		return imageName, imageType, err
	}
	if err != nil {
		s.say(s.client.LastError.Error())
		return "", "", err
	}
	if model := vm.Model; model == nil {
		return "", "", errors.New("TODO")
	}
	if vm.Model.Properties.StorageProfile.OsDisk.Vhd != nil {
		imageType = "image"
		imageName = *vm.Model.Properties.StorageProfile.OsDisk.Vhd.Uri
		return imageType, imageName, nil
	}

	if vm.Model.Properties.StorageProfile.OsDisk.ManagedDisk.Id == nil {
		return "", "", fmt.Errorf("unable to obtain a OS disk for %q, please check that the instance has been created", computeName)
	}

	imageType = "Microsoft.Compute/disks"
	imageName = *vm.Model.Properties.StorageProfile.OsDisk.ManagedDisk.Id

	return imageType, imageName, nil
}

func (s *StepDeployTemplate) Cleanup(state multistep.StateBag) {
	// TODO are there any resources created in DTL builds we should tear down?
	// There was teardown code from the ARM builder copy pasted in but it was never called
}
