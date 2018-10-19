provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "vpc" {
  source = "../../modules/vpc"
  aws_profile = "${var.aws_profile}"
}