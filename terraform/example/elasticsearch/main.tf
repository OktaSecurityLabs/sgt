provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "elasticsearch" {
  source = "../../modules/elasticsearch"
  user_ip_address = "${var.user_ip_address}"
  aws_profile = "${var.aws_profile}"
  create_elasticsearch = "${var.create_elasticsearch}"
}