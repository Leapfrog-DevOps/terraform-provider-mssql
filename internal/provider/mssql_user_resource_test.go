// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMssqlUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMssqlUserResourceConfig("test_user", "master", "test_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_user.test", "name", "test_user"),
					resource.TestCheckResourceAttr("mssql_user.test", "database", "master"),
					resource.TestCheckResourceAttr("mssql_user.test", "login", "test_login"),
					resource.TestCheckResourceAttrSet("mssql_user.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mssql_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccMssqlUserResourceConfig("test_user_updated", "master", "test_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_user.test", "name", "test_user_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMssqlUserResourceConfig(name, database, login string) string {
	return fmt.Sprintf(`
resource "mssql_user" "test" {
  name     = %[1]q
  database = %[2]q
  login    = %[3]q
}
`, name, database, login)
}