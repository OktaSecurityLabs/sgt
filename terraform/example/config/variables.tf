variable "aws_profile" {
}

variable "aws_region" {}

variable "region" {
  default = "us-east-1"
}

variable "full_ssl_certchain" {}
variable "ssl_private_key" {}

variable "terraform_backend_bucket_name" {}

variable "environment" {}