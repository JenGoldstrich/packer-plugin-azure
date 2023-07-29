// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	hashiImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/images"
	hashiVMImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/virtualmachineimages"
	hashiVMSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/virtualmachines"
	hashiDisksSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/disks"
	hashiSnapshotsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/snapshots"
	hashiGalleryImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimages"
	hashiGalleryImageVersionsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
)

var _ AzureClientSet = &AzureClientSetMock{}

// AzureClientSetMock provides a generic mock for AzureClientSet
type AzureClientSetMock struct {
	DisksClientMock                hashiDisksSDK.DisksClient
	SnapshotsClientMock            hashiSnapshotsSDK.SnapshotsClient
	ImagesClientMock               hashiImagesSDK.ImagesClient
	VirtualMachinesClientMock      hashiVMSDK.VirtualMachinesClient
	VirtualMachineImagesClientMock hashiVMImagesSDK.VirtualMachineImagesClient
	GalleryImagesClientMock        hashiGalleryImagesSDK.GalleryImagesClient
	GalleryImageVersionsClientMock hashiGalleryImageVersionsSDK.GalleryImageVersionsClient
	MetadataClientMock             MetadataClientAPI
	SubscriptionIDMock             string
}

// DisksClient returns a DisksClient
func (m *AzureClientSetMock) DisksClient() hashiDisksSDK.DisksClient {
	return m.DisksClientMock
}

// SnapshotsClient returns a SnapshotsClient
func (m *AzureClientSetMock) SnapshotsClient() hashiSnapshotsSDK.SnapshotsClient {
	return m.SnapshotsClientMock
}

// ImagesClient returns a ImagesClient
func (m *AzureClientSetMock) ImagesClient() hashiImagesSDK.ImagesClient {
	return m.ImagesClientMock
}

// VirtualMachineImagesClient returns a VirtualMachineImagesClient
func (m *AzureClientSetMock) VirtualMachineImagesClient() hashiVMImagesSDK.VirtualMachineImagesClient {
	return m.VirtualMachineImagesClientMock
}

// VirtualMachinesClient returns a VirtualMachinesClient
func (m *AzureClientSetMock) VirtualMachinesClient() hashiVMSDK.VirtualMachinesClient {
	return m.VirtualMachinesClientMock
}

// GalleryImagesClient returns a GalleryImagesClient
func (m *AzureClientSetMock) GalleryImagesClient() hashiGalleryImagesSDK.GalleryImagesClient {
	return m.GalleryImagesClientMock
}

// GalleryImageVersionsClient returns a GalleryImageVersionsClient
func (m *AzureClientSetMock) GalleryImageVersionsClient() hashiGalleryImageVersionsSDK.GalleryImageVersionsClient {
	return m.GalleryImageVersionsClientMock
}

// MetadataClient returns a MetadataClient
func (m *AzureClientSetMock) MetadataClient() MetadataClientAPI {
	return m.MetadataClientMock
}

// SubscriptionID returns SubscriptionIDMock
func (m *AzureClientSetMock) SubscriptionID() string {
	return m.SubscriptionIDMock
}
