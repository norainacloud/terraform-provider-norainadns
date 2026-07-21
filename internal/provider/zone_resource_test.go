// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZoneResourceConfig(domainName, "Test zone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_zone.test", "domain_name", domainName),
					resource.TestCheckResourceAttr("norainadns_zone.test", "description", "Test zone"),
					resource.TestCheckResourceAttr("norainadns_zone.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("norainadns_zone.test", "id"),
					resource.TestCheckResourceAttrSet("norainadns_zone.test", "sync_status"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "norainadns_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccZoneResourceConfig(domainName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_zone.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccZoneResource_withNameservers(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneResourceConfig(domainName, "Zone with NS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_zone.test", "domain_name", domainName),
					resource.TestCheckResourceAttrSet("norainadns_zone.test", "nameservers.#"),
				),
			},
		},
	})
}

func testAccZoneResourceConfig(domainName, description string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = %[2]q
}
`, domainName, description)
}
