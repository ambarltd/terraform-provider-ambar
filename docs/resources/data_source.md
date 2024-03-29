---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ambar_data_source Resource - terraform-provider-ambar"
subcategory: ""
description: |-
  Ambar DataSource resource. Represents the details needed for Ambar to establish a connection to your database storage which is then used to import record sequences into Ambar.
---

# ambar_data_source (Resource)

Ambar DataSource resource. Represents the details needed for Ambar to establish a connection to your database storage which is then used to import record sequences into Ambar.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_source_config` (Map of String) A Key Value map of further DataSource configurations specific to the type of database this DataSource will connect to. See Ambar documentation for a list of required parameters.
- `data_source_type` (String) The type of durable storage being connected to. This should be one of the supported database types by Ambar such as postgres. See Ambar documentation for a full list of supported data_source_types.

### Optional

- `description` (String) A user friendly description of this DataSource. Use the description field to help augment information about this DataSource which may not be apparent from describing the resource, such as if it is a test environment resource or which department owns it.

### Read-Only

- `resource_id` (String) The unique Ambar resource id for this resource.
- `state` (String) The current state of the Ambar resource.

## Import

Import is supported using the following syntax:

```shell
# Ambar DataSources can be imported by specifying the resource identifier.
# Note: Some sensitive fields like usernames and passwords will not get imported into Terraform state
# from existing resources and may require further action to manage via Terraform templates.
terraform import ambar_data_source.example_data_source AMBAR-1234567890
```
