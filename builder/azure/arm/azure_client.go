// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arm

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"net/http"

	"github.com/Azure/go-autorest/autorest"
	hashiImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/images"
	hashiVMSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-01/virtualmachines"
	hashiDisksSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/disks"
	hashiSnapshotsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-02/snapshots"
	hashiGalleryImagesSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimages"
	hashiGalleryImageVersionsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/compute/2022-03-03/galleryimageversions"
	hashiSecretsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/keyvault/2023-02-01/secrets"
	hashiVaultsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/keyvault/2023-02-01/vaults"
	hashiNetworkMetaSDK "github.com/hashicorp/go-azure-sdk/resource-manager/network/2022-09-01"
	hashiDeploymentOperationsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/deploymentoperations"
	hashiDeploymentsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/deployments"
	hashiGroupsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/resources/2022-09-01/resourcegroups"
	hashiStorageAccountsSDK "github.com/hashicorp/go-azure-sdk/resource-manager/storage/2022-09-01/storageaccounts"
	authWrapper "github.com/hashicorp/go-azure-sdk/sdk/auth/autorest"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	commonclient "github.com/hashicorp/packer-plugin-azure/builder/azure/common/client"
	"github.com/hashicorp/packer-plugin-azure/version"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
	giovanniBlobStorageSDK "github.com/tombuildsstuff/giovanni/storage/2020-08-04/blob/blobs"
)

const (
	EnvPackerLogAzureMaxLen = "PACKER_LOG_AZURE_MAXLEN"
)

type AzureClient struct {
	NetworkMetaClient hashiNetworkMetaSDK.Client
	hashiDeploymentsSDK.DeploymentsClient
	hashiStorageAccountsSDK.StorageAccountsClient
	hashiDeploymentOperationsSDK.DeploymentOperationsClient
	hashiImagesSDK.ImagesClient
	hashiVMSDK.VirtualMachinesClient
	hashiSecretsSDK.SecretsClient
	hashiVaultsSDK.VaultsClient
	hashiDisksSDK.DisksClient
	hashiGroupsSDK.ResourceGroupsClient
	hashiSnapshotsSDK.SnapshotsClient
	hashiGalleryImageVersionsSDK.GalleryImageVersionsClient
	hashiGalleryImagesSDK.GalleryImagesClient
	GiovanniBlobClient giovanniBlobStorageSDK.Client
	InspectorMaxLength int
	LastError          azureErrorResponse
}

func errorCapture(client *AzureClient) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			body, bodyString := handleBody(resp.Body, math.MaxInt64)
			resp.Body = body

			errorResponse := newAzureErrorResponse(bodyString)
			if errorResponse != nil {
				client.LastError = *errorResponse
			}

			return r.Respond(resp)
		})
	}
}

// WAITING(chrboum): I have logged https://github.com/Azure/azure-sdk-for-go/issues/311 to get this
// method included in the SDK.  It has been accepted, and I'll cut over to the official way
// once it ships.
func byConcatDecorators(decorators ...autorest.RespondDecorator) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.DecorateResponder(r, decorators...)
	}
}

// Returns an Azure Client used for the Azure Resource Manager
// Also returns the Azure object ID for the authentication method used in the build
func NewAzureClient(ctx context.Context, isVHDBuild bool, cloud *environments.Environment, sharedGalleryTimeout time.Duration, pollingDuration time.Duration, newSdkAuthOptions commonclient.NewSDKAuthOptions) (*AzureClient, *string, error) {

	var azureClient = &AzureClient{}

	maxlen := getInspectorMaxLength()
	if cloud == nil || cloud.ResourceManager == nil {
		// TODO Throw error message that helps users solve this problem
		return nil, nil, fmt.Errorf("Azure Environment not configured correctly")
	}
	resourceManagerEndpoint, _ := cloud.ResourceManager.Endpoint()
	resourceManagerAuthorizer, err := commonclient.BuildResourceManagerAuthorizer(ctx, newSdkAuthOptions, *cloud)
	if err != nil {
		return nil, nil, err
	}

	// Clients that have been ported to hashicorp/go-azure-sdk
	azureClient.DisksClient = hashiDisksSDK.NewDisksClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.DisksClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.DisksClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.DisksClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DisksClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.DisksClient.Client.UserAgent)
	azureClient.DisksClient.Client.PollingDuration = pollingDuration

	azureClient.VirtualMachinesClient = hashiVMSDK.NewVirtualMachinesClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.VirtualMachinesClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.VirtualMachinesClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.VirtualMachinesClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VirtualMachinesClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.VirtualMachinesClient.Client.UserAgent)
	azureClient.VirtualMachinesClient.Client.PollingDuration = pollingDuration

	azureClient.SnapshotsClient = hashiSnapshotsSDK.NewSnapshotsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.SnapshotsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.SnapshotsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.SnapshotsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SnapshotsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.SnapshotsClient.Client.UserAgent)
	azureClient.SnapshotsClient.Client.PollingDuration = pollingDuration

	azureClient.SecretsClient = hashiSecretsSDK.NewSecretsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.SecretsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.SecretsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.SecretsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SecretsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.SecretsClient.Client.UserAgent)
	azureClient.SecretsClient.Client.PollingDuration = pollingDuration

	azureClient.VaultsClient = hashiVaultsSDK.NewVaultsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.VaultsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.VaultsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.VaultsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VaultsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.VaultsClient.Client.UserAgent)
	azureClient.VaultsClient.Client.PollingDuration = pollingDuration

	azureClient.DeploymentsClient = hashiDeploymentsSDK.NewDeploymentsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.DeploymentsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.DeploymentsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DeploymentsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.DeploymentsClient.Client.UserAgent)
	azureClient.DeploymentsClient.Client.PollingDuration = pollingDuration

	azureClient.DeploymentOperationsClient = hashiDeploymentOperationsSDK.NewDeploymentOperationsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.DeploymentOperationsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.DeploymentOperationsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentOperationsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DeploymentOperationsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.DeploymentOperationsClient.Client.UserAgent)
	azureClient.DeploymentOperationsClient.Client.PollingDuration = pollingDuration

	azureClient.ResourceGroupsClient = hashiGroupsSDK.NewResourceGroupsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.ResourceGroupsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.ResourceGroupsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.ResourceGroupsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.ResourceGroupsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.ResourceGroupsClient.Client.UserAgent)
	azureClient.ResourceGroupsClient.Client.PollingDuration = pollingDuration

	azureClient.ImagesClient = hashiImagesSDK.NewImagesClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.ImagesClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.ImagesClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.ImagesClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.ImagesClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.ImagesClient.Client.UserAgent)
	azureClient.ImagesClient.Client.PollingDuration = pollingDuration

	azureClient.StorageAccountsClient = hashiStorageAccountsSDK.NewStorageAccountsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.StorageAccountsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.StorageAccountsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.StorageAccountsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.StorageAccountsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.StorageAccountsClient.Client.UserAgent)
	azureClient.StorageAccountsClient.Client.PollingDuration = pollingDuration

	// TODO Request/Response inpectors for Track 2
	networkMetaClient, err := hashiNetworkMetaSDK.NewClientWithBaseURI(cloud.ResourceManager, func(c *resourcemanager.Client) {
		c.Client.Authorizer = resourceManagerAuthorizer
		c.Client.UserAgent = "some-user-agent"
	})

	if err != nil {
		return nil, nil, err
	}
	azureClient.NetworkMetaClient = *networkMetaClient

	azureClient.GalleryImageVersionsClient = hashiGalleryImageVersionsSDK.NewGalleryImageVersionsClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.GalleryImageVersionsClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.GalleryImageVersionsClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImageVersionsClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImageVersionsClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.GalleryImageVersionsClient.Client.UserAgent)
	azureClient.GalleryImageVersionsClient.Client.PollingDuration = sharedGalleryTimeout

	azureClient.GalleryImagesClient = hashiGalleryImagesSDK.NewGalleryImagesClientWithBaseURI(*resourceManagerEndpoint)
	azureClient.GalleryImagesClient.Client.Authorizer = authWrapper.AutorestAuthorizer(resourceManagerAuthorizer)
	azureClient.GalleryImagesClient.Client.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImagesClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImagesClient.Client.UserAgent = fmt.Sprintf("%s %s", useragent.String(version.AzurePluginVersion.FormattedVersion()), azureClient.GalleryImagesClient.Client.UserAgent)
	azureClient.GalleryImagesClient.Client.PollingDuration = pollingDuration

	// We only need the Blob Client to delete the OS VHD during VHD builds
	if isVHDBuild {
		storageAccountAuthorizer, err := commonclient.BuildStorageAuthorizer(ctx, newSdkAuthOptions, *cloud)
		if err != nil {
			return nil, nil, err
		}

		blobClient := giovanniBlobStorageSDK.New()
		azureClient.GiovanniBlobClient = blobClient
		azureClient.GiovanniBlobClient.Authorizer = authWrapper.AutorestAuthorizer(storageAccountAuthorizer)
		azureClient.GiovanniBlobClient.Client.RequestInspector = withInspection(maxlen)
		azureClient.GiovanniBlobClient.Client.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	}

	token, err := resourceManagerAuthorizer.Token(ctx, &http.Request{})
	if err != nil {
		return nil, nil, err
	}
	// TODO Handle potential panic here if Access Token or child objects are null
	objectId, err := commonclient.GetObjectIdFromToken(token.AccessToken)
	if err != nil {
		return nil, nil, err
	}
	return azureClient, &objectId, nil
}

func getInspectorMaxLength() int64 {
	value, ok := os.LookupEnv(EnvPackerLogAzureMaxLen)
	if !ok {
		return math.MaxInt64
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	if i < 0 {
		return 0
	}

	return i
}
