
variable "arm_client_id" {
  type    = string
  default = "${env("ARM_CLIENT_ID")}"
}

variable "arm_oidc_token" {
  type    = string
  default = "${env("ARM_OIDC_TOKEN")}"
}

variable "subscription_id" {
  type    = string
  default = "${env("ARM_SUBSCRIPTION_ID")}"
}

source "azure-arm" "autogenerated_1" {
  client_id                         = "${var.arm_client_id}"
  client_jwt                        = "${var.arm_oidc_token}"
  communicator                      = "winrm"
  image_offer                       = "WindowsServer"
  image_publisher                   = "MicrosoftWindowsServer"
  image_sku                         = "2012-R2-Datacenter"
  location                          = "South Central US"
  managed_image_name                = "oidc-example"
  managed_image_resource_group_name = "packer-acceptance-test"
  os_type                           = "Windows"
  subscription_id                   = "${var.subscription_id}"
  vm_size                           = "Standard_DS2_v2"
  winrm_insecure                    = "true"
  winrm_timeout                     = "3m"
  winrm_use_ssl                     = "true"
  winrm_username                    = "packer"
}

build {
  sources = ["source.azure-arm.autogenerated_1"]

}
