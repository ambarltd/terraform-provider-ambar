resource "ambar_filter" "example_filter" {
  data_source_id  = ambar_data_source.example_data_source.resource_id
  description     = "My test Filter"
  filter_contents = ""
}