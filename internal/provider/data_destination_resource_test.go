package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	exampleDataDestinationResourceConfig = `
resource "ambar_data_destination" "test_destination" {
  filter_ids = [
    ambar_filter.test_filter.resource_id
  ]
  description = "My Terraform DataDestination"
  destination_endpoint = "https://1.2.3.4.com/data"
  destination_name = "ambar-dest"
  username = "username"
  password = "password"
}`
)

func TestAmbarDataDestinationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				// Filters require that there is a DataSource to use for type checking.
				Config: providerConfig + exampleDataSourceConfig + exampleFilterResourceConfig + exampleDataDestinationResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify placeholder id attribute
					resource.TestCheckResourceAttrSet("ambar_filter.test_filter", "resource_id"),
				),
			},
		},
	})
}
