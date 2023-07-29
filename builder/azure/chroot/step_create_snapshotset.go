// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	hashiSnapshotsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/snapshots"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/Azure/go-autorest/autorest/to"
)

var _ multistep.Step = &StepCreateSnapshotset{}

type StepCreateSnapshotset struct {
	OSDiskSnapshotID         string
	DataDiskSnapshotIDPrefix string
	Location                 string

	SkipCleanup bool

	snapshots Diskset
}

func (s *StepCreateSnapshotset) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	diskset := state.Get(stateBagKey_Diskset).(Diskset)

	s.snapshots = make(Diskset)

	errorMessage := func(format string, params ...interface{}) multistep.StepAction {
		err := fmt.Errorf("StepCreateSnapshotset.Run: error: "+format, params...)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for lun, resource := range diskset {
		snapshotID := fmt.Sprintf("%s%d", s.DataDiskSnapshotIDPrefix, lun)
		if lun == -1 {
			snapshotID = s.OSDiskSnapshotID
		}
		ssr, err := client.ParseResourceID(snapshotID)
		if err != nil {
			errorMessage("Could not create a valid resource id, tried %q: %v", snapshotID, err)
		}
		if !strings.EqualFold(ssr.Provider, "Microsoft.Compute") ||
			!strings.EqualFold(ssr.ResourceType.String(), "snapshots") {
			return errorMessage("Resource %q is not of type Microsoft.Compute/snapshots", snapshotID)
		}
		s.snapshots[lun] = ssr
		state.Put(stateBagKey_Snapshotset, s.snapshots)

		ui.Say(fmt.Sprintf("Creating snapshot %q", ssr))

		resourceID := resource.String()
		snapshot := hashiSnapshotsSDK.Snapshot{
			Location: s.Location,
			Properties: &hashiSnapshotsSDK.SnapshotProperties{
				CreationData: hashiSnapshotsSDK.CreationData{
					CreateOption:     hashiSnapshotsSDK.DiskCreateOptionCopy,
					SourceResourceId: &resourceID,
				},
				Incremental: to.BoolPtr(false),
			},
		}
		snapshotSDKID := hashiSnapshotsSDK.NewSnapshotID(azcli.SubscriptionID(), ssr.ResourceGroup, ssr.ResourceName.String())
		err = azcli.SnapshotsClient().CreateOrUpdateThenPoll(ctx, snapshotSDKID, snapshot)
		if err != nil {
			return errorMessage("error initiating snapshot %q: %v", ssr, err)
		}

	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshotset) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packersdk.Ui)

		for _, resource := range s.snapshots {

			snapshotID := hashiSnapshotsSDK.NewSnapshotID(azcli.SubscriptionID(), resource.ResourceGroup, resource.ResourceName.String())
			ui.Say(fmt.Sprintf("Removing any active SAS for snapshot %q", resource))
			{
				err := azcli.SnapshotsClient().RevokeAccessThenPoll(context.TODO(), snapshotID)
				if err != nil {
					log.Printf("StepCreateSnapshotset.Cleanup: error: %+v", err)
					ui.Error(fmt.Sprintf("error deleting snapshot %q: %v.", resource, err))
				}
			}

			ui.Say(fmt.Sprintf("Deleting snapshot %q", resource))
			{
				err := azcli.SnapshotsClient().DeleteThenPoll(context.TODO(), snapshotID)
				if err != nil {
					log.Printf("StepCreateSnapshotset.Cleanup: error: %+v", err)
					ui.Error(fmt.Sprintf("error deleting snapshot %q: %v.", resource, err))
				}
			}
		}
	}
}
