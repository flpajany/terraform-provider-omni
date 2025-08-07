# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    omni = {
      version = "0.1.0"
      source  = "flpajany/omni"
    }
  }
}

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

data "omni_talosconfig" "talosconfig" {
  cluster_name = "omni-cluster-1"
}
