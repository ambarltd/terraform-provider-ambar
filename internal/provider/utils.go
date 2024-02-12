package provider

import (
	"regexp"
	"strings"
)

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// Small utility functions to help convert the usual camel case used in HTTP
// to the names we give fields in Terraform. This is just to remove some cognitive overhead
// for Terraform users.
func toSnakeCase(s string) string {
	snake := matchAllCap.ReplaceAllString(s, "${1}_${2}")
	return snake
}

// AmbarApiErrorToTerraformErrorString Extracts out just the error portion of the JSON body of the http response.
// We then use the other util functions to clean it up and correct the naming to be correct as fields would appear
// in Terraform template files.
func AmbarApiErrorToTerraformErrorString(apiString string) string {
	slicedString := strings.SplitAfter(apiString, "\":")
	errorContent := strings.Trim(slicedString[len(slicedString)-1], "\"{}")
	return toSnakeCase(errorContent)
}
