// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOmniClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOmniClusterResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("omni_cluster.test", "template", `
kind: Cluster
name: test-cluster-1
kubernetes:
  version: "v1.27.12"
talos:
  version: "v1.6.8"`),
				),
			},
			// ImportState testing
			{
				ResourceName:      "omni_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"template", "name"},
			},
			// Update and Read testing
			{
				Config: testAccOmniClusterResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("omni_cluster.test", "template", `
kind: Cluster
name: test-cluster-1
kubernetes:
  version: "v1.29.9"
talos:
  version: "v1.6.8"`),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOmniClusterResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "omni_cluster" "test" {
  hardware_address = %[1]q
}
`, configurableAttribute)
}
