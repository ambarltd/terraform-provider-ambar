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
	# partitioning_column = "partition"
	# serial_column = "serial"
	# username = "postgres"
	# password = "password"
	data_source_config = {
		# "hostname": "host",
		"hostPort": "5432",
		"databaseName": "postgres",
		"tableName": "events",
		"publicationName": "acceptance_test_pub",
		# "additionalColumns": "seqid,seqnum,value"
	}
}`
)

func TestAmbarDataSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				// DataSource just requires a valid provider configuration
				Config: providerConfig + exampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_type", "postgres"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "description", "My Terraform DataSource"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "partitioning_column", "partition"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "serial_column", "serial"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "username", "username"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "password", "password"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "state", "READY"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.hostname", "host"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.hostPort", "5432"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.databaseName", "postgres"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.tableName", "events"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.publicationName", "example_pub"),
					resource.TestCheckResourceAttr("ambar_data_source.test_data_source", "data_source_config.additionalColumns", "some,other,column"),
					// Verify placeholder id attribute
					resource.TestCheckResourceAttrSet("ambar_data_source.test_data_source", "resource_id"),
				),
			},
		},
	})
}
