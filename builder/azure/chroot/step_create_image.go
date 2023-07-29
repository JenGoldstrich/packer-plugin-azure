// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	hashiImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/images"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

var _ multistep.Step = &StepCreateImage{}

type StepCreateImage struct {
	ImageResourceID            string
	ImageOSState               string
	OSDiskStorageAccountType   string
	OSDiskCacheType            string
	DataDiskStorageAccountType string
	DataDiskCacheType          string
	Location                   string
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	diskset := state.Get(stateBagKey_Diskset).(Diskset)
	diskResourceID := diskset.OS().String()

	ui.Say(fmt.Sprintf("Creating image %s\n   using %s for os disk.",
		s.ImageResourceID,
		diskResourceID))

	imageResource, err := azure.ParseResourceID(s.ImageResourceID)

	if err != nil {
		log.Printf("StepCreateImage.Run: error: %+v", err)
		err := fmt.Errorf(
			"error parsing image resource id '%s': %v", s.ImageResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	storageAccountType := hashiImagesSDK.StorageAccountTypes(s.OSDiskStorageAccountType)
	cacheingType := hashiImagesSDK.CachingTypes(s.OSDiskCacheType)
	image := hashiImagesSDK.Image{
		Location: s.Location,
		Properties: &hashiImagesSDK.ImageProperties{
			StorageProfile: &hashiImagesSDK.ImageStorageProfile{
				OsDisk: &hashiImagesSDK.ImageOSDisk{
					OsState: hashiImagesSDK.OperatingSystemStateTypes(s.ImageOSState),
					OsType:  hashiImagesSDK.OperatingSystemTypesLinux,
					ManagedDisk: &hashiImagesSDK.SubResource{
						Id: &diskResourceID,
					},
					StorageAccountType: &storageAccountType,
					Caching:            &cacheingType,
				},
			},
		},
	}

	var datadisks []hashiImagesSDK.ImageDataDisk
	if len(diskset) > 0 {
		storageAccountType = hashiImagesSDK.StorageAccountTypes(s.DataDiskStorageAccountType)
		cacheingType = hashiImagesSDK.CachingTypes(s.DataDiskStorageAccountType)
	}
	for lun, resource := range diskset {
		if lun != -1 {
			ui.Say(fmt.Sprintf("   using %q for data disk (lun %d).", resource, lun))

			datadisks = append(datadisks, hashiImagesSDK.ImageDataDisk{
				Lun:                lun,
				ManagedDisk:        &hashiImagesSDK.SubResource{Id: to.StringPtr(resource.String())},
				StorageAccountType: &storageAccountType,
				Caching:            &cacheingType,
			})
		}
	}
	if datadisks != nil {
		sort.Slice(datadisks, func(i, j int) bool {
			return datadisks[i].Lun < datadisks[j].Lun
		})
		image.Properties.StorageProfile.DataDisks = &datadisks
	}

	id := hashiImagesSDK.NewImageID(azcli.SubscriptionID(), imageResource.ResourceGroup, imageResource.ResourceName)
	err = azcli.ImagesClient().CreateOrUpdateThenPoll(
		ctx,
		id,
		image)
	if err != nil {
		log.Printf("StepCreateImage.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating image '%s': %v", s.ImageResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Image creation complete")

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(bag multistep.StateBag) {} // this is the final artifact, don't delete
