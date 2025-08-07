# Copyright (c) HashiCorp, Inc.

#
# Set TF_VAR_service_account
#
variable "service_account" {
  type = string
}

#
# Set TF_VAR_omni_uri
#
variable "omni_uri" {
  type = string
}

provider "omni" {
  uri             = var.omni_uri
  service_account = var.service_account
}
