// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	hashiDisksSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/disks"

	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests assume current machine is capable of running chroot builder (i.e. an Azure VM)

func Test_DiskAttacherAttachesDiskToVM(t *testing.T) {
	azcli, err := client.GetTestClientSet(t) // integration test
	require.Nil(t, err)
	testDiskName := t.Name()

	errorBuffer := &strings.Builder{}
	ui := &packersdk.BasicUi{
		Reader:      strings.NewReader(""),
		Writer:      ioutil.Discard,
		ErrorWriter: errorBuffer,
	}

	da := NewDiskAttacher(azcli, ui)

	vm, err := azcli.MetadataClient().GetComputeInfo()
	require.Nil(t, err, "Test needs to run on an Azure VM, unable to retrieve VM information")
	t.Log("Creating new disk '", testDiskName, "' in ", vm.ResourceGroupName)

	diskId := hashiDisksSDK.NewDiskID(azcli.SubscriptionID(), vm.ResourceGroupName, testDiskName)

	disk, err := azcli.DisksClient().Get(context.TODO(), diskId)
	if err == nil {
		t.Log("Disk already exists")
		if *disk.Model.Properties.DiskState == hashiDisksSDK.DiskStateAttached {
			t.Log("Disk is attached, assuming to this machine, trying to detach")
			err = da.DetachDisk(context.TODO(), *disk.Model.Id)
			require.Nil(t, err)
		}
		t.Log("Deleting disk")
		err := azcli.DisksClient().DeleteThenPoll(context.TODO(), diskId)
		require.Nil(t, err)
	}

	t.Log("Creating disk")
	var diskSizeInGb int64
	diskSizeInGb = 30
	diskSkuName := hashiDisksSDK.DiskStorageAccountTypesStandardLRS
	err = azcli.DisksClient().CreateOrUpdateThenPoll(context.TODO(), diskId,
		hashiDisksSDK.Disk{
			Location: vm.Location,
			Sku: &hashiDisksSDK.DiskSku{
				Name: &diskSkuName,
			},
			Properties: &hashiDisksSDK.DiskProperties{
				DiskSizeGB:   &diskSizeInGb,
				CreationData: hashiDisksSDK.CreationData{CreateOption: hashiDisksSDK.DiskCreateOptionEmpty},
			},
		})
	require.Nil(t, err)

	t.Log("Retrieving disk properties")
	d, err := azcli.DisksClient().Get(context.TODO(), diskId)
	require.Nil(t, err)
	assert.NotNil(t, d)

	t.Log("Attaching disk")
	lun, err := da.AttachDisk(context.TODO(), *d.Model.Id)
	assert.Nil(t, err)

	t.Log("Waiting for device")
	dev, err := da.WaitForDevice(context.TODO(), lun)
	assert.Nil(t, err)

	t.Log("Device path:", dev)

	t.Log("Detaching disk")
	err = da.DetachDisk(context.TODO(), *d.Model.Id)
	require.Nil(t, err)

	t.Log("Deleting disk")
	err = azcli.DisksClient().DeleteThenPoll(context.TODO(), diskId)
	require.Nil(t, err)
}
