package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	exampleDataSourceConfig = `
resource "ambar_data_source" "test_data_source" {
	data_source_type = "postgres"
	description = "My Terraform Acceptance Test DataSource"
	data_source_config = {
		"hostname": "hostname",
		"hostPort": "5432",
		"databaseName": "postgres",
		"tableName": "events",
		"publicationName": "acceptance_test_pub",
		"columns": "partitioning_column,serial_column,columns",
		"partitioning_column = "partitioning_column"
		"serial_column = "serial_column"
		"username = "username"
		"password = "password"
	}
}`
)

func TestAccAmbarDataSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				// DataSource just requires a valid provider configuration
				Config: providerConfig + exampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify placeholder id attribute
					resource.TestCheckResourceAttrSet("ambar_data_source.test_data_source", "resource_id"),
				),
			},
		},
	})
}
