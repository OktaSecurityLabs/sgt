variable "sgt-s3-osquery-results-bucket-name" {}

variable "aws_profile" {}

variable "aws_region" {
  default = "us-east-1"
}

variable "create_elasticsearch" {
  default = 1
}