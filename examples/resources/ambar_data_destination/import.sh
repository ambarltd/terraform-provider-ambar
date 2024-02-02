# Ambar DataDestinations can be imported by specifying the resource identifier.
# Note: Some sensitive fields like usernames and passwords will not get imported into Terraform state
# from existing resources and may require further action to manage via Terraform templates.
terraform import ambar_data_destination.example_data_destination AMBAR-1234567890