resource "ambar_data_source" "example_data_source" {
  data_source_type    = "postgres"
  description         = "My Terraform DataSource"
  partitioning_column = "partition"
  serial_column       = "serial"
  username            = "username"
  password            = "password"
  data_source_config = {
    "hostname" : "host",
    "hostPort" : "5432",
    "databaseName" : "postgres",
    "tableName" : "events",
    "publicationName" : "example_pub",
    "additionalColumns" : "some,other,column"
  }
}