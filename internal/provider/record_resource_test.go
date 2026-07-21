// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRecordResource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRecordResourceConfig_basic(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "name", "www"),
					resource.TestCheckResourceAttr("norainadns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("norainadns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("norainadns_record.test", "routing_policy", "simple"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.#", "1"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.0.value", "192.0.2.1"),
					resource.TestCheckResourceAttrSet("norainadns_record.test", "id"),
					resource.TestCheckResourceAttrSet("norainadns_record.test", "fqdn"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "norainadns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRecordImportStateIdFunc("norainadns_record.test"),
			},
			// Update testing
			{
				Config: testAccRecordResourceConfig_updated(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "ttl", "600"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.0.value", "192.0.2.2"),
				),
			},
		},
	})
}

func TestAccRecordResource_cname(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_cname(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "name", "blog"),
					resource.TestCheckResourceAttr("norainadns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.0.value", "www.example.com"),
				),
			},
		},
	})
}

func TestAccRecordResource_txt(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_txt(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "name", "_dmarc"),
					resource.TestCheckResourceAttr("norainadns_record.test", "type", "TXT"),
				),
			},
		},
	})
}

func TestAccRecordResource_mx(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_mx(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "name", "@"),
					resource.TestCheckResourceAttr("norainadns_record.test", "type", "MX"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.#", "2"),
				),
			},
		},
	})
}

func TestAccRecordResource_failover(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	domainName := fmt.Sprintf("%s.example.com", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecordResourceConfig_failover(domainName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("norainadns_record.test", "name", "app"),
					resource.TestCheckResourceAttr("norainadns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("norainadns_record.test", "routing_policy", "failover"),
					resource.TestCheckResourceAttr("norainadns_record.test", "is_dynamic", "true"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.#", "2"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.0.priority", "0"),
					resource.TestCheckResourceAttr("norainadns_record.test", "values.1.priority", "1"),
					resource.TestCheckResourceAttrSet("norainadns_record.test", "values.0.monitor_id"),
				),
			},
		},
	})
}

func testAccRecordImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Resource not found: %s", resourceName)
		}

		zoneID := rs.Primary.Attributes["zone_id"]
		recordID := rs.Primary.ID

		return fmt.Sprintf("%s/%s", zoneID, recordID), nil
	}
}

func testAccRecordResourceConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for records"
}

resource "norainadns_record" "test" {
  zone_id = norainadns_zone.test.id
  name    = "www"
  type    = "A"
  ttl     = 300

  values {
    value = "192.0.2.1"
  }
}
`, domainName)
}

func testAccRecordResourceConfig_updated(domainName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for records"
}

resource "norainadns_record" "test" {
  zone_id = norainadns_zone.test.id
  name    = "www"
  type    = "A"
  ttl     = 600

  values {
    value = "192.0.2.2"
  }
}
`, domainName)
}

func testAccRecordResourceConfig_cname(domainName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for CNAME"
}

resource "norainadns_record" "test" {
  zone_id = norainadns_zone.test.id
  name    = "blog"
  type    = "CNAME"
  ttl     = 300

  values {
    value = "www.example.com"
  }
}
`, domainName)
}

func testAccRecordResourceConfig_txt(domainName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for TXT"
}

resource "norainadns_record" "test" {
  zone_id = norainadns_zone.test.id
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600

  values {
    value = "v=DMARC1; p=reject; rua=mailto:dmarc@example.com"
  }
}
`, domainName)
}

func testAccRecordResourceConfig_mx(domainName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for MX"
}

resource "norainadns_record" "test" {
  zone_id = norainadns_zone.test.id
  name    = "@"
  type    = "MX"
  ttl     = 3600

  values {
    value    = "10 mail1.example.com."
    priority = 0
  }

  values {
    value    = "20 mail2.example.com."
    priority = 1
  }
}
`, domainName)
}

func testAccRecordResourceConfig_failover(domainName, monitorName string) string {
	return fmt.Sprintf(`
resource "norainadns_zone" "test" {
  domain_name = %[1]q
  description = "Test zone for failover"
}

resource "norainadns_monitor" "test" {
  name        = %[2]q
  protocol    = "HTTPS"
  target_host = "origin.example.com"
  target_path = "/health"
}

resource "norainadns_record" "test" {
  zone_id        = norainadns_zone.test.id
  name           = "app"
  type           = "A"
  ttl            = 60
  is_dynamic     = true
  routing_policy = "failover"

  values {
    value      = "192.0.2.10"
    priority   = 0
    monitor_id = norainadns_monitor.test.id
  }

  values {
    value    = "192.0.2.20"
    priority = 1
  }
}
`, domainName, monitorName)
}
