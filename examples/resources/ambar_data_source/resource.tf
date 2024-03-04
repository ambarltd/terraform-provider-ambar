resource "ambar_data_source" "example_data_source" {
  data_source_type = "postgres"
  description      = "My Terraform Postgres DataSource"
  # data_source_config key-values depend on the type of DataSource being created.
  # See Ambar docs for more details.
  data_source_config = {
    "hostname" : "host",
    "hostPort" : 5432,
    "username" : "username",
    "password" : "password"
    "databaseName" : "postgres",
    "tableName" : "events",
    "publicationName" : "example_pub",
    "partitioningColumn" : "partition",
    "serialColumn" : "serial",
    # columns should include all columns to be read from the database
    # including the partition and serial columns
    "columns" : "partition,serial,some,other,column",
    # tls termination override is optional
    "tlsTerminationOverrideHost" : "tls.termination.host"
  }
}

resource "ambar_data_source" "example_mysql_data_source" {
  data_source_type = "mysql"
  description      = "My Terraform MySQL DataSource"
  # data_source_config key-values depend on the type of DataSource being created.
  # See Ambar docs for more details.
  data_source_config = {
    "hostname" : "host",
    "hostPort" : 3036,
    "username" : "username",
    "password" : "password"
    "databaseName" : "mysql",
    "tableName" : "events",
    "partitioningColumn" : "partition",
    "incrementingColumn" : "incrementing",
    # columns should include all columns to be read from the database
    # including the partition and incrementing columns
    "columns" : "partition,incrementing,some,other,column",
    "binLogReplicationServerId" : 1001,
    # tls termination override is optional
    "tlsTerminationOverrideHost" : "tls.termination.host"
  }
}