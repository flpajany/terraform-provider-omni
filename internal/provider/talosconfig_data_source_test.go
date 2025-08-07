// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOmniTalosconfigDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccOmniTalosconfigDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.omni_talosconfig.test", "cluster_name", "omni-cluster-1"),
				),
			},
		},
	})
}

const testAccOmniTalosconfigDataSourceConfig = `
data "omni_talosconfig" "test" {
  cluster_name = "omni-cluster-1"
}
`
