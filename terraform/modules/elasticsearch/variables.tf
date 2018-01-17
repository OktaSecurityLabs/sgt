variable "user_ip_address" {}
variable "aws_profile" {}
variable "aws_region" {
  default = "us-east-1"
}

variable "elasticsearch_domain_name" {
  default = "sgt-osquery-results"
}