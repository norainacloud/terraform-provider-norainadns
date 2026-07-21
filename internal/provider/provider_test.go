// Copyright (c) Noraina
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"norainadns": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates the necessary test API keys exist in the environment.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NORAINADNS_ENDPOINT"); v == "" {
		t.Fatal("NORAINADNS_ENDPOINT must be set for acceptance tests")
	}
	if v := os.Getenv("NORAINADNS_API_KEY"); v == "" {
		t.Fatal("NORAINADNS_API_KEY must be set for acceptance tests")
	}
}
