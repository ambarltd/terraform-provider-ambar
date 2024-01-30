terraform {
  required_providers {
    ambar = {
      source = "ambarltd/ambar"
    }
  }
}

provider "ambar" {
  endpoint = "region.api.ambar.cloud"
  api_key  = "your-key"
}
