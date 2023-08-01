// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	hashiGalleryImageVersionsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestStepCreateSharedImageVersion_Run(t *testing.T) {
	standardZRSStorageType := hashiGalleryImageVersionsSDK.StorageAccountTypeStandardZRS
	hostCacheingRW := hashiGalleryImageVersionsSDK.HostCachingReadWrite
	hostCacheingNone := hashiGalleryImageVersionsSDK.HostCachingNone
	subscriptionID := "12345"
	type fields struct {
		Destination       SharedImageGalleryDestination
		OSDiskCacheType   string
		DataDiskCacheType string
		Location          string
	}
	tests := []struct {
		name                 string
		fields               fields
		snapshotset          Diskset
		want                 multistep.StepAction
		expectedImageVersion hashiGalleryImageVersionsSDK.GalleryImageVersion
		expectedImageId      hashiGalleryImageVersionsSDK.ImageVersionId
	}{
		{
			name: "happy path",
			fields: fields{
				Destination: SharedImageGalleryDestination{
					ResourceGroup: "ResourceGroup",
					GalleryName:   "GalleryName",
					ImageName:     "ImageName",
					ImageVersion:  "0.1.2",
					TargetRegions: []TargetRegion{
						{
							Name:               "region1",
							ReplicaCount:       5,
							StorageAccountType: "Standard_ZRS",
						},
					},
					ExcludeFromLatest: true,
				},
				OSDiskCacheType:   "ReadWrite",
				DataDiskCacheType: "None",
				Location:          "region2",
			},
			snapshotset: diskset(
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/osdisksnapshot",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot0",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot1",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot2"),
			expectedImageId: hashiGalleryImageVersionsSDK.NewImageVersionID(
				subscriptionID,
				"ResourceGroup",
				"GalleryName",
				"ImageName",
				"0.1.2",
			),
			expectedImageVersion: hashiGalleryImageVersionsSDK.GalleryImageVersion{
				Location: "region2",
				Properties: &hashiGalleryImageVersionsSDK.GalleryImageVersionProperties{
					PublishingProfile: &hashiGalleryImageVersionsSDK.GalleryArtifactPublishingProfileBase{
						ExcludeFromLatest: common.BoolPtr(true),
						TargetRegions: &[]hashiGalleryImageVersionsSDK.TargetRegion{
							{
								Name:                 "region1",
								RegionalReplicaCount: common.Int64Ptr(5),
								StorageAccountType:   &standardZRSStorageType,
							},
						},
					},
					StorageProfile: hashiGalleryImageVersionsSDK.GalleryImageVersionStorageProfile{
						OsDiskImage: &hashiGalleryImageVersionsSDK.GalleryDiskImage{
							Source: &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{
								Id: common.StringPtr("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/osdisksnapshot"),
							},
							HostCaching: &hostCacheingRW,
						},
						DataDiskImages: &[]hashiGalleryImageVersionsSDK.GalleryDataDiskImage{
							{
								HostCaching: &hostCacheingNone,
								Lun:         0,
								Source: &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{
									Id: common.StringPtr("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot0"),
								},
							},
							{
								HostCaching: &hostCacheingNone,
								Lun:         1,
								Source: &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{
									Id: common.StringPtr("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot1"),
								},
							},
							{
								HostCaching: &hostCacheingNone,
								Lun:         2,
								Source: &hashiGalleryImageVersionsSDK.GalleryDiskImageSource{
									Id: common.StringPtr("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot2"),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		state := new(multistep.BasicStateBag)
		state.Put("azureclient", &client.AzureClientSetMock{
			SubscriptionIDMock: subscriptionID,
		})
		state.Put("ui", packersdk.TestUi(t))
		state.Put(stateBagKey_Snapshotset, tt.snapshotset)

		t.Run(tt.name, func(t *testing.T) {
			var actualID hashiGalleryImageVersionsSDK.ImageVersionId
			var actualImageVersion hashiGalleryImageVersionsSDK.GalleryImageVersion
			s := &StepCreateSharedImageVersion{
				Destination:       tt.fields.Destination,
				OSDiskCacheType:   tt.fields.OSDiskCacheType,
				DataDiskCacheType: tt.fields.DataDiskCacheType,
				Location:          tt.fields.Location,
				create: func(ctx context.Context, azcli client.AzureClientSet, id hashiGalleryImageVersionsSDK.ImageVersionId, imageVersion hashiGalleryImageVersionsSDK.GalleryImageVersion) error {
					actualID = id
					actualImageVersion = imageVersion
					return nil
				},
			}

			action := s.Run(context.TODO(), state)
			if action != multistep.ActionContinue {
				t.Fatalf("Expected ActionContinue got %s", action)
			}
			if diff := cmp.Diff(actualImageVersion, tt.expectedImageVersion); diff != "" {
				t.Fatalf("unexpected image version %s", diff)
			}
			if actualID != tt.expectedImageId {
				t.Fatalf("Expected image ID %+v got %+v", tt.expectedImageId, actualID)
			}
		})
	}
}
