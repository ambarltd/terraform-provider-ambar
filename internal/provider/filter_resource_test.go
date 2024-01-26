package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	exampleFilterResourceConfig = `
resource "ambar_filter" "test_filter" {
  data_source_id = ambar_data_source.test_data_source.resource_id
  description = "My test Filter"
  filter_contents = ""
}`
)

func TestAmbarFilterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				// Filters require that there is a DataSource to use for type checking.
				Config: providerConfig + exampleDataSourceConfig + exampleFilterResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify placeholder id attribute
					resource.TestCheckResourceAttrSet("ambar_filter.test_filter", "resource_id"),
				),
			},
		},
	})
}
