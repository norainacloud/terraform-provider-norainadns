// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMonitorResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMonitorResourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_monitor.test", "name", rName),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "target_host", "example.com"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "target_path", "/health"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("norainadns_monitor.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "norainadns_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccMonitorResourceConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_monitor.test", "name", rName+"-updated"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "target_path", "/healthz"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "timeout_seconds", "15"),
				),
			},
		},
	})
}

func TestAccMonitorResource_withForceIP(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorResourceConfig_withForceIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_monitor.test", "name", rName),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "force_ip_address", "192.0.2.50"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "custom_host_header", "origin.example.com"),
				),
			},
		},
	})
}

func TestAccMonitorResource_http(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMonitorResourceConfig_http(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_monitor.test", "name", rName),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("norainadns_monitor.test", "target_port", "8080"),
				),
			},
		},
	})
}

func testAccMonitorResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "norainadns_monitor" "test" {
  name        = %[1]q
  protocol    = "HTTPS"
  target_host = "example.com"
  target_path = "/health"
}
`, name)
}

func testAccMonitorResourceConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "norainadns_monitor" "test" {
  name            = "%[1]s-updated"
  protocol        = "HTTPS"
  target_host     = "example.com"
  target_path     = "/healthz"
  timeout_seconds = 15
}
`, name)
}

func testAccMonitorResourceConfig_withForceIP(name string) string {
	return fmt.Sprintf(`
resource "norainadns_monitor" "test" {
  name              = %[1]q
  protocol          = "HTTPS"
  target_host       = "cdn.example.com"
  target_path       = "/health"
  force_ip_address  = "192.0.2.50"
  custom_host_header = "origin.example.com"
}
`, name)
}

func testAccMonitorResourceConfig_http(name string) string {
	return fmt.Sprintf(`
resource "norainadns_monitor" "test" {
  name        = %[1]q
  protocol    = "HTTP"
  target_host = "example.com"
  target_port = 8080
  target_path = "/status"
}
`, name)
}
