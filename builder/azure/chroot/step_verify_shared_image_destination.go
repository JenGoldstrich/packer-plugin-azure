// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"
	hashiGalleryImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimages"
	hashiGalleryImageVersionsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
	"github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

var _ multistep.Step = &StepVerifySharedImageDestination{}

// StepVerifySharedImageDestination verifies that the shared image location matches the Location field in the step.
// Also verifies that the OS Type is Linux.
type StepVerifySharedImageDestination struct {
	Image    SharedImageGalleryDestination
	Location string
}

// Run retrieves the image metadata from Azure and compares the location to Location. Verifies the OS Type.
func (s *StepVerifySharedImageDestination) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)

	errorMessage := func(message string, parameters ...interface{}) multistep.StepAction {
		err := fmt.Errorf(message, parameters...)
		log.Printf("StepVerifySharedImageDestination.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imageURI := fmt.Sprintf("/subscriptions/%s/resourcegroup/%s/providers/Microsoft.Compute/galleries/%s/images/%s",
		azcli.SubscriptionID(),
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName,
	)

	ui.Say(fmt.Sprintf("Validating that shared image %s exists", imageURI))
	galleryImageID := hashiGalleryImagesSDK.NewGalleryImageID(
		azcli.SubscriptionID(),
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName,
	)
	imageResult, err := azcli.GalleryImagesClient().Get(ctx, galleryImageID)

	if err != nil {
		return errorMessage("Error retrieving shared image %q: %+v ", imageURI, err)
	}

	image := imageResult.Model

	if image.Id == nil || *image.Id == "" {
		return errorMessage("Error retrieving shared image %q: ID field in response is empty", imageURI)
	}
	if image.Properties == nil {
		return errorMessage("Could not retrieve shared image properties for image %q.", image.Id)
	}

	location := image.Location

	log.Printf("StepVerifySharedImageDestination:Run: Image %q, Location: %q, HvGen: %q, osState: %q",
		to.String(image.Id),
		location,
		image.Properties.HyperVGeneration,
		image.Properties.OsState)

	if !strings.EqualFold(location, s.Location) {
		return errorMessage("Destination shared image resource %q is in a different location (%q) than this VM (%q). "+
			"Packer does not know how to handle that.",
			to.String(image.Id),
			location,
			s.Location)
	}

	if image.Properties.OsType != hashiGalleryImagesSDK.OperatingSystemTypesLinux {
		return errorMessage("The shared image (%q) is not a Linux image (found %q). Currently only Linux images are supported.",
			to.String(image.Id),
			image.Properties.OsType)
	}

	ui.Say(fmt.Sprintf("Found image %s in location %s",
		*image.Id,
		image.Location,
	))

	// TODO Suggest moving gallery image ID to common IDs library
	// so we don't have to define two different versions of the same resource ID
	galleryImageIDForList := hashiGalleryImageVersionsSDK.NewGalleryImageID(
		azcli.SubscriptionID(),
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName,
	)
	versions, err := azcli.GalleryImageVersionsClient().ListByGalleryImageComplete(ctx,
		galleryImageIDForList)

	if err != nil {
		return errorMessage("Could not ListByGalleryImageComplete group:%v gallery:%v image:%v",
			s.Image.ResourceGroup, s.Image.GalleryName, s.Image.ImageName)
	}

	for _, version := range versions.Items {
		if version.Name == nil {
			return errorMessage("Could not retrieve versions for image %q: unexpected nil name", image.Id)
		}
		if *version.Name == s.Image.ImageVersion {
			return errorMessage("Shared image version %q already exists for image %q.", s.Image.ImageVersion, image.Id)
		}
	}

	return multistep.ActionContinue
}

func (*StepVerifySharedImageDestination) Cleanup(multistep.StateBag) {}
