resource "ambar_data_source" "example_data_source" {
  data_source_type    = "postgres"
  description         = "My Terraform DataSource"
  partitioning_column = "partition"
  serial_column       = "serial"
  username            = "username"
  password            = "password"
  # data_source_config key-values depend on the type of DataSource being created.
  # See Ambar docs for more details.
  data_source_config = {
    "hostname" : "host",
    "hostPort" : "5432",
    "databaseName" : "postgres",
    "tableName" : "events",
    "publicationName" : "example_pub",
    # columns should include all columns to be read from the database
    # including the partition and serial columns
    "columns" : "partition,serial,some,other,column"
  }
}