variable "user_ip_address" {}
variable "aws_profile" {}
variable "aws_region" {
  default = "us-east-1"
}

variable "user_pool_name" {
  default = "kibana_user_pool"
}
variable "identity_pool_name" {
  default = "kibana_identity_pool"
}
variable "es_cognito_access_role_name" {
  default = "es_cognito_role"
}

variable "elasticsearch_domain_name" {
  default = "sgt-osquery-results"
}

variable "create_elasticsearch" {
  default = 1
  description = "toggles the creation of elasticsearch"
}
