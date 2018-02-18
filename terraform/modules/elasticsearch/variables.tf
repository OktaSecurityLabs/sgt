variable "user_ip_address" {}
variable "aws_profile" {}
variable "aws_region" {
  default = "us-east-1"
}

variable "elasticsearch_domain_name" {
  default = "sgt-osquery-results"
}

variable "create_elasticsearch" {
  default = 1
  description = "toggles the creation of elasticsearch"
}