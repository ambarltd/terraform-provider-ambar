terraform {
  required_providers {
    ambar = {
      source = "ambar.cloud/terraform/ambar"
    }
  }
}

provider "ambar" {
  endpoint = "region.api.ambar.cloud"
  api_key  = "your-key"
}
