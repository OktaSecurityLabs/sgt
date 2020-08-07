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

variable "elasticsearch_instance_count" {
  default = 3
}

variable "elasticsearch_instance_type" {
  default = "m4.large.elasticsearch"
}

variable "elasticsearch_dedicated_master_count" {
  default = 3
}

variable "elasticsearch_master_instance_type" {
  default = "t2.medium.elasticsearch"
}

variable "elasticsearch_volume_size" {
  default = 300
}

variable "elasticsearch_volume_type" {
  default = "gp2"
}
