// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arm

// these tests require the following variables to be set,
// although some test will only use a subset:
//
// * ARM_CLIENT_ID
// * ARM_CLIENT_SECRET
// * ARM_SUBSCRIPTION_ID
// * ARM_STORAGE_ACCOUNT
//
// The subscription in question should have a resource group
// called "packer-acceptance-test" in "South Central US" region. The
// storage account referred to in the above variable should
// be inside this resource group and in "South Central US" as well.
//
// There should be a shared image gallery inside of the resource group
// it should be called `acctestgallery` in "South Central US" as well.
//
// In addition, the PACKER_ACC variable should also be set to
// a non-empty value to enable Packer acceptance tests and the
// options "-v -timeout 90m" should be provided to the test
// command, e.g.:
//   go test -v -timeout 90m -run TestBuilderAcc_.*

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimages"
	"github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
	commonclient "github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

func TestBuilderAcc_SharedImageGallery_ARM64SpecializedLinuxSIG(t *testing.T) {
	t.Parallel()
	if os.Getenv("AZURE_CLI_AUTH") == "" {
		t.Skip("Azure CLI Acceptance tests skipped unless env 'AZURE_CLI_AUTH' is set, and an active `az login` session has been established")
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-specialized-linux-sig",
		Type:     "azure-arm",
		Template: string(armLinuxSpecialziedSIGTemplate),
		Setup: func() error {
			createSharedImageGalleryDefinition(t, CreateSharedImageGalleryDefinitionParameters{
				galleryImageName: "arm-linux-specialized-sig",
				subscriptionId:   os.Getenv("ARM_SUBSCRIPTION_ID"),
				imageSku:         "22_04-lts-arm64",
				imageOffer:       "0001-com-ubuntu-server-jammy",
				imagePublisher:   "canonical",
				isX64:            false,
				isWindows:        false,
				useGenTwoVM:      true,
				specialized:      true,
			})
			return nil
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
		Teardown: func() error {
			deleteSharedImageGalleryDefinition(t, os.Getenv("ARM_SUBSCRIPTION_ID"), "arm-linux-specialized-sig")
			return nil
		},
	})
}

func TestBuilderAcc_SharedImageGallery_WindowsSIG(t *testing.T) {
	t.Parallel()
	if os.Getenv("AZURE_CLI_AUTH") == "" {
		t.Skip("Azure CLI Acceptance tests skipped unless env 'AZURE_CLI_AUTH' is set, and an active `az login` session has been established")
		return
	}
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-windows-sig",
		Type:     "azure-arm",
		Template: string(windowsSIGTemplate),
		Setup: func() error {
			createSharedImageGalleryDefinition(t, CreateSharedImageGalleryDefinitionParameters{
				galleryImageName: "windows-sig",
				imageSku:         "2012-R2-Datacenter",
				subscriptionId:   os.Getenv("ARM_SUBSCRIPTION_ID"),
				imageOffer:       "WindowsServer",
				imagePublisher:   "MicrosoftWindowsServer",
				isX64:            true,
				isWindows:        true,
			})
			return nil
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
		Teardown: func() error {
			deleteSharedImageGalleryDefinition(t, os.Getenv("ARM_SUBSCRIPTION_ID"), "windows-sig")
			return nil
		},
	})
}

func TestBuilderAcc_ManagedDisk_Windows(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-managedisk-windows",
		Type:     "azure-arm",
		Template: testBuilderAccManagedDiskWindows,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_ManagedDisk_Windows_Build_Resource_Group(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-managedisk-windows-build-resource-group",
		Type:     "azure-arm",
		Template: testBuilderAccManagedDiskWindowsBuildResourceGroup,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_ManagedDisk_Windows_Build_Resource_Group_Additional_Disk(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-managedisk-windows-build-resource-group-additional-disk",
		Type:     "azure-arm",
		Template: testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionalDisk,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_ManagedDisk_Linux(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-managedisk-linux",
		Type:     "azure-arm",
		Template: testBuilderAccManagedDiskLinux,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_ManagedDisk_Linux_AzureCLI(t *testing.T) {
	t.Parallel()
	if os.Getenv("AZURE_CLI_AUTH") == "" {
		t.Skip("Azure CLI Acceptance tests skipped unless env 'AZURE_CLI_AUTH' is set, and an active `az login` session has been established")
		return
	}

	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-managedisk-linux-azurecli",
		Type:     "azure-arm",
		Template: testBuilderAccManagedDiskLinuxAzureCLI,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_Blob_Windows(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-blob-windows",
		Type:     "azure-arm",
		Template: testBuilderAccBlobWindows,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderAcc_Blob_Linux(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-blob-linux",
		Type:     "azure-arm",
		Template: testBuilderAccBlobLinux,
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

func TestBuilderUserData_Linux(t *testing.T) {
	t.Parallel()
	tmpfile, err := ioutil.TempFile("", "userdata")
	if err != nil {
		t.Fatalf("failed creating tempfile: %s", err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(testBuilderCustomDataLinux); err != nil {
		t.Fatalf("failed writing userdata: %s", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed closing file: %s", err)
	}

	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-userdata-linux",
		Type:     "azure-arm",
		Template: testBuilderUserDataLinux(tmpfile.Name()),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

//go:embed testdata/rsa_sha2_only_server.pkr.hcl
var rsaSHA2OnlyTemplate []byte

//go:embed testdata/windows_sig.pkr.hcl
var windowsSIGTemplate []byte

//go:embed testdata/arm_linux_specialized.pkr.hcl
var armLinuxSpecialziedSIGTemplate []byte

func TestBuilderAcc_rsaSHA2OnlyServer(t *testing.T) {
	t.Parallel()
	acctest.TestPlugin(t, &acctest.PluginTestCase{
		Name:     "test-azure-ubuntu-jammy-linux",
		Type:     "azure-arm",
		Template: string(rsaSHA2OnlyTemplate),
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	})
}

type CreateSharedImageGalleryDefinitionParameters struct {
	galleryImageName string
	subscriptionId   string
	imageSku         string
	imageOffer       string
	imagePublisher   string
	isX64            bool
	isWindows        bool
	useGenTwoVM      bool
	specialized      bool
}

func createTestAzureClient(t *testing.T) AzureClient {
	b := Builder{}
	_, _, _ = b.Prepare()
	ui := testUi()
	// Use CLI auth for our test client
	b.config.ClientConfig.UseAzureCLIAuth = true
	_ = b.config.ClientConfig.FillParameters()
	authOptions := commonclient.NewSDKAuthOptions{
		AuthType:       b.config.ClientConfig.AuthType(),
		ClientID:       b.config.ClientConfig.ClientID,
		ClientSecret:   b.config.ClientConfig.ClientSecret,
		TenantID:       b.config.ClientConfig.TenantID,
		SubscriptionID: b.config.ClientConfig.SubscriptionID,
	}
	ui.Message("Creating test Azure Resource Manager (ARM) client ...")
	azureClient, _, err := NewAzureClient(
		context.TODO(),
		true,
		b.config.ClientConfig.CloudEnvironment(),
		b.config.SharedGalleryTimeout,
		b.config.PollingDurationTimeout,
		authOptions)
	if err != nil {
		t.Fatalf("failed to create test azure client: %s", err)
	}
	return *azureClient
}

func createSharedImageGalleryDefinition(t *testing.T, params CreateSharedImageGalleryDefinitionParameters) {
	azureClient := createTestAzureClient(t)
	osType := galleryimages.OperatingSystemTypesLinux
	if params.isWindows {
		osType = galleryimages.OperatingSystemTypesWindows
	}
	osState := galleryimages.OperatingSystemStateTypesGeneralized
	if params.specialized {
		osState = galleryimages.OperatingSystemStateTypesSpecialized
	}
	osArch := galleryimages.ArchitectureArmSixFour
	if params.isX64 {
		osArch = galleryimages.ArchitectureXSixFour
	}
	hyperVGeneration := galleryimages.HyperVGenerationVOne
	if params.useGenTwoVM {
		hyperVGeneration = galleryimages.HyperVGenerationVTwo
	}
	location := "southcentralus"
	galleryId := galleryimages.NewGalleryImageID(params.subscriptionId, "packer-acceptance-test", "acctestgallery", params.galleryImageName)
	_, err := azureClient.GalleryImagesClient.CreateOrUpdate(context.TODO(), galleryId, galleryimages.GalleryImage{
		Properties: &galleryimages.GalleryImageProperties{
			OsType:           osType,
			OsState:          osState,
			Architecture:     &osArch,
			HyperVGeneration: &hyperVGeneration,
			Identifier: galleryimages.GalleryImageIdentifier{
				Publisher: params.imagePublisher,
				Offer:     params.imageOffer,
				Sku:       params.imageSku,
			},
		},
		Location: location,
	})

	if err != nil {
		t.Fatalf("failed to create Gallery %s: %s", params.galleryImageName, err)
	}
}

func deleteSharedImageGalleryDefinition(t *testing.T, subscriptionId, galleryImageName string) {
	azureClient := createTestAzureClient(t)
	galleryVersionId := galleryimageversions.NewImageVersionID(subscriptionId, "packer-acceptance-test", "acctestgallery", galleryImageName, "1.0.0")
	err := azureClient.GalleryImageVersionsClient.DeleteThenPoll(context.TODO(), galleryVersionId)
	if err != nil {
		t.Fatalf("failed to delete Gallery %s: %s", galleryImageName, err)
	}
	retryConfig := retry.Config{
		Tries:      5,
		RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}
	err = retryConfig.Run(context.TODO(), func(ctx context.Context) error {
		galleryId := galleryimages.NewGalleryImageID(subscriptionId, "packer-acceptance-test", "acctestgallery", galleryImageName)
		err := azureClient.GalleryImagesClient.DeleteThenPoll(context.TODO(), galleryId)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to delete Gallery %s: %s", galleryImageName, err)
	}
}

func testUi() *packersdk.BasicUi {
	return &packersdk.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}
}

func testBuilderUserDataLinux(userdata string) string {
	return fmt.Sprintf(`
{
	"variables": {
	  "client_id": "{{env `+"`ARM_CLIENT_ID`"+`}}",
	  "client_secret": "{{env `+"`ARM_CLIENT_SECRET`"+`}}",
	  "subscription_id": "{{env `+"`ARM_SUBSCRIPTION_ID`"+`}}",
	  "storage_account": "{{env `+"`ARM_STORAGE_ACCOUNT`"+`}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user `+"`client_id`"+`}}",
	  "client_secret": "{{user `+"`client_secret`"+`}}",
	  "subscription_id": "{{user `+"`subscription_id`"+`}}",

	  "storage_account": "{{user `+"`storage_account`"+`}}",
	  "resource_group_name": "packer-acceptance-test",
	  "capture_container_name": "test",
	  "capture_name_prefix": "testBuilderUserDataLinux",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",
	  "user_data_file": "%s",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`, userdata)
}

const testBuilderCustomDataLinux = `#cloud-config
growpart:
  mode: off
`

const testBuilderAccManagedDiskWindows = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindows-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",
	  "async_resourcegroup_delete": "true",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskWindowsBuildResourceGroup = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "build_resource_group_name" : "packer-acceptance-test",
	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsBuildResourceGroup-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",
	  "async_resourcegroup_delete": "true",

	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionalDisk = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "build_resource_group_name" : "packer-acceptance-test",
	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsBuildResourceGroupAdditionDisk-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",
	  "async_resourcegroup_delete": "true",

	  "vm_size": "Standard_DS2_v2",
	  "disk_additional_size": [10,15]
	}]
}
`

const testBuilderAccManagedDiskWindowsDeviceLogin = `
{
	"variables": {
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskWindowsDeviceLogin-{{timestamp}}",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskLinux = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinux-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2",
	  "azure_tags": {
	    "env": "testing",
	    "builder": "packer"
	   }
	}]
}
`

const testBuilderAccManagedDiskLinuxDeviceLogin = `
{
	"variables": {
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinuxDeviceLogin-{{timestamp}}",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",
	  "async_resourcegroup_delete": "true",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccBlobWindows = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "storage_account": "{{env ` + "`ARM_STORAGE_ACCOUNT`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "storage_account": "{{user ` + "`storage_account`" + `}}",
	  "resource_group_name": "packer-acceptance-test",
	  "capture_container_name": "azure-arm",
	  "capture_name_prefix": "testBuilderAccBlobWin",

	  "os_type": "Windows",
	  "image_publisher": "MicrosoftWindowsServer",
	  "image_offer": "WindowsServer",
	  "image_sku": "2012-R2-Datacenter",

	  "communicator": "winrm",
	  "winrm_use_ssl": "true",
	  "winrm_insecure": "true",
	  "winrm_timeout": "3m",
	  "winrm_username": "packer",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccBlobLinux = `
{
	"variables": {
	  "client_id": "{{env ` + "`ARM_CLIENT_ID`" + `}}",
	  "client_secret": "{{env ` + "`ARM_CLIENT_SECRET`" + `}}",
	  "subscription_id": "{{env ` + "`ARM_SUBSCRIPTION_ID`" + `}}",
	  "storage_account": "{{env ` + "`ARM_STORAGE_ACCOUNT`" + `}}"
	},
	"builders": [{
	  "type": "azure-arm",

	  "client_id": "{{user ` + "`client_id`" + `}}",
	  "client_secret": "{{user ` + "`client_secret`" + `}}",
	  "subscription_id": "{{user ` + "`subscription_id`" + `}}",

	  "storage_account": "{{user ` + "`storage_account`" + `}}",
	  "resource_group_name": "packer-acceptance-test",
	  "capture_container_name": "test",
	  "capture_name_prefix": "testBuilderAccBlobLinux",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2"
	}]
}
`

const testBuilderAccManagedDiskLinuxAzureCLI = `
{
	"builders": [{
	  "type": "azure-arm",

	  "use_azure_cli_auth": true,

	  "managed_image_resource_group_name": "packer-acceptance-test",
	  "managed_image_name": "testBuilderAccManagedDiskLinuxAzureCLI-{{timestamp}}",
	  "temp_resource_group_name": "packer-acceptance-test-managed-cli",

	  "os_type": "Linux",
	  "image_publisher": "Canonical",
	  "image_offer": "UbuntuServer",
	  "image_sku": "16.04-LTS",

	  "location": "South Central US",
	  "vm_size": "Standard_DS2_v2",
	  "azure_tags": {
	    "env": "testing",
	    "builder": "packer"
	   }
	}]
}
`
