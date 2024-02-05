## 1.0.1
FEATURES:
* Removed DataDestination DestinationName field
* Removed DataSource top level fields, they should instead be passed as part of the DataSourceConfig map
* Updated provider to use the latest Ambar SDK
* Minor improvements to debug logging

## 1.0.0 (Initial Release)
FEATURES:
 * Ambar initial Terraform support.
   * Support for Ambar DataSource resources like the Postgres DataSourceType
   * Support for Ambar Filter resources, allowing to define a record sequence filter to be applied to a DataSource.
   * Support for Ambar DataDestination resources, allowing delivery of one or more filtered record sequences from one or more DataDestinations