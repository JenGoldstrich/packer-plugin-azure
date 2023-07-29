// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/Azure/go-autorest/autorest/to"
	hashiGalleryImageVersionsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCreateSharedImageVersion struct {
	Destination       SharedImageGalleryDestination
	OSDiskCacheType   string
	DataDiskCacheType string
	Location          string
}

func (s *StepCreateSharedImageVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	snapshotset := state.Get(stateBagKey_Snapshotset).(Diskset)

	ui.Say(fmt.Sprintf("Creating image version %s\n   using %q for os disk.",
		s.Destination.ResourceID(azcli.SubscriptionID()),
		snapshotset.OS()))

	var targetRegions []hashiGalleryImageVersionsSDK.TargetRegion
	// transform target regions to API objects
	for _, tr := range s.Destination.TargetRegions {
		trStorageAccountType := hashiGalleryImageVersionsSDK.StorageAccountType(tr.StorageAccountType)
		apiObject := hashiGalleryImageVersionsSDK.TargetRegion{
			Name:                 tr.Name,
			RegionalReplicaCount: &tr.ReplicaCount,
			StorageAccountType:   &trStorageAccountType,
		}
		targetRegions = append(targetRegions, apiObject)
	}

	osDiskSource := snapshotset.OS().String()
	hostCaching := hashiGalleryImageVersionsSDK.HostCaching(s.OSDiskCacheType)
	imageVersion := hashiGalleryImageVersionsSDK.GalleryImageVersion{
		Location: s.Location,
		Properties: &hashiGalleryImageVersionsSDK.GalleryImageVersionProperties{
			StorageProfile: hashiGalleryImageVersionsSDK.GalleryImageVersionStorageProfile{
				OsDiskImage: &hashiGalleryImageVersionsSDK.GalleryDiskImage{
					Source:      &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{Id: &osDiskSource},
					HostCaching: &hostCaching,
				},
			},
			PublishingProfile: &hashiGalleryImageVersionsSDK.GalleryArtifactPublishingProfileBase{
				TargetRegions:     &targetRegions,
				ExcludeFromLatest: to.BoolPtr(s.Destination.ExcludeFromLatest),
			},
		},
	}

	var datadisks []hashiGalleryImageVersionsSDK.GalleryDataDiskImage
	for lun, resource := range snapshotset {
		if lun != -1 {
			ui.Say(fmt.Sprintf("   using %q for data disk (lun %d).", resource, lun))

			hostCaching := hashiGalleryImageVersionsSDK.HostCaching(s.DataDiskCacheType)
			datadisks = append(datadisks, hashiGalleryImageVersionsSDK.GalleryDataDiskImage{
				Lun:         lun,
				Source:      &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{Id: common.StringPtr(resource.String())},
				HostCaching: &hostCaching,
			})
		}
	}
	if datadisks != nil {
		// sort by lun
		sort.Slice(datadisks, func(i, j int) bool {
			return datadisks[i].Lun < datadisks[j].Lun
		})
		imageVersion.Properties.StorageProfile.DataDiskImages = &datadisks
	}

	galleryImageVersionID := hashiGalleryImageVersionsSDK.NewImageVersionID(
		azcli.SubscriptionID(),
		s.Destination.ResourceGroup,
		s.Destination.GalleryName,
		s.Destination.ImageName,
		s.Destination.ImageVersion,
	)
	err := azcli.GalleryImageVersionsClient().CreateOrUpdateThenPoll(
		ctx,
		galleryImageVersionID,
		imageVersion)
	if err != nil {
		log.Printf("StepCreateSharedImageVersion.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating shared image version '%s': %v", s.Destination.ResourceID(azcli.SubscriptionID()), err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Image creation complete")

	return multistep.ActionContinue
}

func (*StepCreateSharedImageVersion) Cleanup(multistep.StateBag) {}
