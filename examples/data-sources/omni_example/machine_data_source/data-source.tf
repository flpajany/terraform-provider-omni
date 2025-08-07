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

data "omni_machine" "machine1" {
  uuid = "954d3242-0d82-0cb3-f8de-7d3342c2051d"
}

data "omni_machine" "machine2" {
  hardware_address      = "00:50:56:b2:6a:03" // this MAC address should be known from omni
  wait_for_registration = true
}

data "omni_machine" "example3" {
  hardware_address      = "00:50:56:b2:6a:05" // for testing, this MAC address should be unknown from omni
  wait_for_registration = true                // this option at true will make terraform wait till provider timeout until the MAC address is know from omni
}