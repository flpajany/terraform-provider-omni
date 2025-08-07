// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOmniKubeconfigResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOmniKubeconfigResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("omni_kubeconfig.test", "cluster_name", "omni-cluster-1")),
			},
			// ImportState testing
			{
				ResourceName:      "omni_kubeconfig.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"cluster_name"},
			},
			// Update and Read testing
			{
				Config: testAccOmniKubeconfigResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("omni_kubeconfig.test", "cluster_name", "omni-cluster-1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOmniKubeconfigResourceConfig() string {
	return ` resource "omni_kubeconfig" "test" {
  cluster_name = "omni-cluster-1" } `
}
