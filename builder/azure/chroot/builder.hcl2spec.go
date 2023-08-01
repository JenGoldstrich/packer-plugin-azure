// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package chroot

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	PackerBuildName                   *string                            `mapstructure:"packer_build_name" cty:"packer_build_name" hcl:"packer_build_name"`
	PackerBuilderType                 *string                            `mapstructure:"packer_builder_type" cty:"packer_builder_type" hcl:"packer_builder_type"`
	PackerCoreVersion                 *string                            `mapstructure:"packer_core_version" cty:"packer_core_version" hcl:"packer_core_version"`
	PackerDebug                       *bool                              `mapstructure:"packer_debug" cty:"packer_debug" hcl:"packer_debug"`
	PackerForce                       *bool                              `mapstructure:"packer_force" cty:"packer_force" hcl:"packer_force"`
	PackerOnError                     *string                            `mapstructure:"packer_on_error" cty:"packer_on_error" hcl:"packer_on_error"`
	PackerUserVars                    map[string]string                  `mapstructure:"packer_user_variables" cty:"packer_user_variables" hcl:"packer_user_variables"`
	PackerSensitiveVars               []string                           `mapstructure:"packer_sensitive_variables" cty:"packer_sensitive_variables" hcl:"packer_sensitive_variables"`
	SkipCreateImage                   *bool                              `mapstructure:"skip_create_image" required:"false" cty:"skip_create_image" hcl:"skip_create_image"`
	CloudEnvironmentName              *string                            `mapstructure:"cloud_environment_name" required:"false" cty:"cloud_environment_name" hcl:"cloud_environment_name"`
	MetadataHost                      *string                            `mapstructure:"metadata_host" required:"false" cty:"metadata_host" hcl:"metadata_host"`
	ClientID                          *string                            `mapstructure:"client_id" cty:"client_id" hcl:"client_id"`
	ClientSecret                      *string                            `mapstructure:"client_secret" cty:"client_secret" hcl:"client_secret"`
	ClientCertPath                    *string                            `mapstructure:"client_cert_path" cty:"client_cert_path" hcl:"client_cert_path"`
	ClientCertExpireTimeout           *string                            `mapstructure:"client_cert_token_timeout" required:"false" cty:"client_cert_token_timeout" hcl:"client_cert_token_timeout"`
	ClientJWT                         *string                            `mapstructure:"client_jwt" cty:"client_jwt" hcl:"client_jwt"`
	ObjectID                          *string                            `mapstructure:"object_id" cty:"object_id" hcl:"object_id"`
	TenantID                          *string                            `mapstructure:"tenant_id" required:"false" cty:"tenant_id" hcl:"tenant_id"`
	SubscriptionID                    *string                            `mapstructure:"subscription_id" cty:"subscription_id" hcl:"subscription_id"`
	UseAzureCLIAuth                   *bool                              `mapstructure:"use_azure_cli_auth" required:"false" cty:"use_azure_cli_auth" hcl:"use_azure_cli_auth"`
	FromScratch                       *bool                              `mapstructure:"from_scratch" cty:"from_scratch" hcl:"from_scratch"`
	Source                            *string                            `mapstructure:"source" required:"true" cty:"source" hcl:"source"`
	CommandWrapper                    *string                            `mapstructure:"command_wrapper" cty:"command_wrapper" hcl:"command_wrapper"`
	PreMountCommands                  []string                           `mapstructure:"pre_mount_commands" cty:"pre_mount_commands" hcl:"pre_mount_commands"`
	MountOptions                      []string                           `mapstructure:"mount_options" cty:"mount_options" hcl:"mount_options"`
	MountPartition                    *string                            `mapstructure:"mount_partition" cty:"mount_partition" hcl:"mount_partition"`
	MountPath                         *string                            `mapstructure:"mount_path" cty:"mount_path" hcl:"mount_path"`
	PostMountCommands                 []string                           `mapstructure:"post_mount_commands" cty:"post_mount_commands" hcl:"post_mount_commands"`
	ChrootMounts                      [][]string                         `mapstructure:"chroot_mounts" cty:"chroot_mounts" hcl:"chroot_mounts"`
	CopyFiles                         []string                           `mapstructure:"copy_files" cty:"copy_files" hcl:"copy_files"`
	OSDiskSizeGB                      *int64                             `mapstructure:"os_disk_size_gb" cty:"os_disk_size_gb" hcl:"os_disk_size_gb"`
	OSDiskStorageAccountType          *string                            `mapstructure:"os_disk_storage_account_type" cty:"os_disk_storage_account_type" hcl:"os_disk_storage_account_type"`
	OSDiskCacheType                   *string                            `mapstructure:"os_disk_cache_type" cty:"os_disk_cache_type" hcl:"os_disk_cache_type"`
	DataDiskStorageAccountType        *string                            `mapstructure:"data_disk_storage_account_type" cty:"data_disk_storage_account_type" hcl:"data_disk_storage_account_type"`
	DataDiskCacheType                 *string                            `mapstructure:"data_disk_cache_type" cty:"data_disk_cache_type" hcl:"data_disk_cache_type"`
	ImageHyperVGeneration             *string                            `mapstructure:"image_hyperv_generation" cty:"image_hyperv_generation" hcl:"image_hyperv_generation"`
	TemporaryOSDiskID                 *string                            `mapstructure:"temporary_os_disk_id" cty:"temporary_os_disk_id" hcl:"temporary_os_disk_id"`
	TemporaryOSDiskSnapshotID         *string                            `mapstructure:"temporary_os_disk_snapshot_id" cty:"temporary_os_disk_snapshot_id" hcl:"temporary_os_disk_snapshot_id"`
	TemporaryDataDiskIDPrefix         *string                            `mapstructure:"temporary_data_disk_id_prefix" cty:"temporary_data_disk_id_prefix" hcl:"temporary_data_disk_id_prefix"`
	TemporaryDataDiskSnapshotIDPrefix *string                            `mapstructure:"temporary_data_disk_snapshot_id" cty:"temporary_data_disk_snapshot_id" hcl:"temporary_data_disk_snapshot_id"`
	SkipCleanup                       *bool                              `mapstructure:"skip_cleanup" cty:"skip_cleanup" hcl:"skip_cleanup"`
	ImageResourceID                   *string                            `mapstructure:"image_resource_id" cty:"image_resource_id" hcl:"image_resource_id"`
	SharedImageGalleryDestination     *FlatSharedImageGalleryDestination `mapstructure:"shared_image_destination" cty:"shared_image_destination" hcl:"shared_image_destination"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"packer_build_name":               &hcldec.AttrSpec{Name: "packer_build_name", Type: cty.String, Required: false},
		"packer_builder_type":             &hcldec.AttrSpec{Name: "packer_builder_type", Type: cty.String, Required: false},
		"packer_core_version":             &hcldec.AttrSpec{Name: "packer_core_version", Type: cty.String, Required: false},
		"packer_debug":                    &hcldec.AttrSpec{Name: "packer_debug", Type: cty.Bool, Required: false},
		"packer_force":                    &hcldec.AttrSpec{Name: "packer_force", Type: cty.Bool, Required: false},
		"packer_on_error":                 &hcldec.AttrSpec{Name: "packer_on_error", Type: cty.String, Required: false},
		"packer_user_variables":           &hcldec.AttrSpec{Name: "packer_user_variables", Type: cty.Map(cty.String), Required: false},
		"packer_sensitive_variables":      &hcldec.AttrSpec{Name: "packer_sensitive_variables", Type: cty.List(cty.String), Required: false},
		"skip_create_image":               &hcldec.AttrSpec{Name: "skip_create_image", Type: cty.Bool, Required: false},
		"cloud_environment_name":          &hcldec.AttrSpec{Name: "cloud_environment_name", Type: cty.String, Required: false},
		"metadata_host":                   &hcldec.AttrSpec{Name: "metadata_host", Type: cty.String, Required: false},
		"client_id":                       &hcldec.AttrSpec{Name: "client_id", Type: cty.String, Required: false},
		"client_secret":                   &hcldec.AttrSpec{Name: "client_secret", Type: cty.String, Required: false},
		"client_cert_path":                &hcldec.AttrSpec{Name: "client_cert_path", Type: cty.String, Required: false},
		"client_cert_token_timeout":       &hcldec.AttrSpec{Name: "client_cert_token_timeout", Type: cty.String, Required: false},
		"client_jwt":                      &hcldec.AttrSpec{Name: "client_jwt", Type: cty.String, Required: false},
		"object_id":                       &hcldec.AttrSpec{Name: "object_id", Type: cty.String, Required: false},
		"tenant_id":                       &hcldec.AttrSpec{Name: "tenant_id", Type: cty.String, Required: false},
		"subscription_id":                 &hcldec.AttrSpec{Name: "subscription_id", Type: cty.String, Required: false},
		"use_azure_cli_auth":              &hcldec.AttrSpec{Name: "use_azure_cli_auth", Type: cty.Bool, Required: false},
		"from_scratch":                    &hcldec.AttrSpec{Name: "from_scratch", Type: cty.Bool, Required: false},
		"source":                          &hcldec.AttrSpec{Name: "source", Type: cty.String, Required: false},
		"command_wrapper":                 &hcldec.AttrSpec{Name: "command_wrapper", Type: cty.String, Required: false},
		"pre_mount_commands":              &hcldec.AttrSpec{Name: "pre_mount_commands", Type: cty.List(cty.String), Required: false},
		"mount_options":                   &hcldec.AttrSpec{Name: "mount_options", Type: cty.List(cty.String), Required: false},
		"mount_partition":                 &hcldec.AttrSpec{Name: "mount_partition", Type: cty.String, Required: false},
		"mount_path":                      &hcldec.AttrSpec{Name: "mount_path", Type: cty.String, Required: false},
		"post_mount_commands":             &hcldec.AttrSpec{Name: "post_mount_commands", Type: cty.List(cty.String), Required: false},
		"chroot_mounts":                   &hcldec.AttrSpec{Name: "chroot_mounts", Type: cty.List(cty.List(cty.String)), Required: false},
		"copy_files":                      &hcldec.AttrSpec{Name: "copy_files", Type: cty.List(cty.String), Required: false},
		"os_disk_size_gb":                 &hcldec.AttrSpec{Name: "os_disk_size_gb", Type: cty.Number, Required: false},
		"os_disk_storage_account_type":    &hcldec.AttrSpec{Name: "os_disk_storage_account_type", Type: cty.String, Required: false},
		"os_disk_cache_type":              &hcldec.AttrSpec{Name: "os_disk_cache_type", Type: cty.String, Required: false},
		"data_disk_storage_account_type":  &hcldec.AttrSpec{Name: "data_disk_storage_account_type", Type: cty.String, Required: false},
		"data_disk_cache_type":            &hcldec.AttrSpec{Name: "data_disk_cache_type", Type: cty.String, Required: false},
		"image_hyperv_generation":         &hcldec.AttrSpec{Name: "image_hyperv_generation", Type: cty.String, Required: false},
		"temporary_os_disk_id":            &hcldec.AttrSpec{Name: "temporary_os_disk_id", Type: cty.String, Required: false},
		"temporary_os_disk_snapshot_id":   &hcldec.AttrSpec{Name: "temporary_os_disk_snapshot_id", Type: cty.String, Required: false},
		"temporary_data_disk_id_prefix":   &hcldec.AttrSpec{Name: "temporary_data_disk_id_prefix", Type: cty.String, Required: false},
		"temporary_data_disk_snapshot_id": &hcldec.AttrSpec{Name: "temporary_data_disk_snapshot_id", Type: cty.String, Required: false},
		"skip_cleanup":                    &hcldec.AttrSpec{Name: "skip_cleanup", Type: cty.Bool, Required: false},
		"image_resource_id":               &hcldec.AttrSpec{Name: "image_resource_id", Type: cty.String, Required: false},
		"shared_image_destination":        &hcldec.BlockSpec{TypeName: "shared_image_destination", Nested: hcldec.ObjectSpec((*FlatSharedImageGalleryDestination)(nil).HCL2Spec())},
	}
	return s
}
