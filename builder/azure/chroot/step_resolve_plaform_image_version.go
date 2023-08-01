// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	hashiVMImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/virtualmachineimages"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepResolvePlatformImageVersion resolves the exact PIR version when the version is 'latest'
type StepResolvePlatformImageVersion struct {
	*client.PlatformImage
	ResourceGroupName string
	Location          string
	list              func(context.Context, client.AzureClientSet, hashiVMImagesSDK.SkuId, hashiVMImagesSDK.ListOperationOptions) (*[]hashiVMImagesSDK.VirtualMachineImageResource, error)
}

func NewStepResolvePlatformImageVersion(step *StepResolvePlatformImageVersion) *StepResolvePlatformImageVersion {
	step.list = step.listVMImages
	return step
}

// Run retrieves all available versions of a PIR image and stores the latest in the PlatformImage
func (pi *StepResolvePlatformImageVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if strings.EqualFold(pi.Version, "latest") {
		azcli := state.Get("azureclient").(client.AzureClientSet)

		//vmi, err := azcli.VirtualMachineImagesClient().GetLatest(ctx, pi.Publisher, pi.Offer, pi.Sku, pi.Location)
		vmMachineImagesSKUID := hashiVMImagesSDK.NewSkuID(azcli.SubscriptionID(), pi.Location, pi.Publisher, pi.Offer, pi.Sku)
		orderBy := "name desc"
		vmList, err := pi.list(
			ctx,
			azcli,
			vmMachineImagesSKUID,
			hashiVMImagesSDK.ListOperationOptions{
				Orderby: &orderBy,
			},
		)
		if err != nil {
			log.Printf("StepResolvePlatformImageVersion.Run: error: %+v", err)
			err := fmt.Errorf("error retieving latest version of %q: %v", pi.URN(), err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if len(*vmList) == 0 {
			err := fmt.Errorf("%s:%s:%s:latest could not be found in location %s", pi.Publisher, pi.Offer, pi.Sku, pi.Location)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		pi.Version = (*vmList)[0].Name
		ui.Say("Resolved latest version of source image: " + pi.Version)
	} else {
		ui.Say("Nothing to do, version is not 'latest'")
	}

	return multistep.ActionContinue
}
func (s *StepResolvePlatformImageVersion) listVMImages(ctx context.Context, azcli client.AzureClientSet, skuID hashiVMImagesSDK.SkuId, operations hashiVMImagesSDK.ListOperationOptions) (*[]hashiVMImagesSDK.VirtualMachineImageResource, error) {
	result, err := azcli.VirtualMachineImagesClient().List(
		ctx,
		skuID,
		operations,
	)
	if err != nil {
		return nil, err
	}
	if result.Model == nil {
		return nil, fmt.Errorf("SDK returned empty model")
	}
	return result.Model, nil
}

func (*StepResolvePlatformImageVersion) Cleanup(multistep.StateBag) {}
