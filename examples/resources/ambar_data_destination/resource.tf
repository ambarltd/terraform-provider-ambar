resource "ambar_data_destination" "example_destination" {
  filter_ids = [
    ambar_filter.example_filter.resource_id
  ]
  description          = "My Terraform DataDestination"
  destination_endpoint = "https://1.2.3.4.com/data"
  destination_port     = "8443"
  username             = "username"
  password             = "password"
}