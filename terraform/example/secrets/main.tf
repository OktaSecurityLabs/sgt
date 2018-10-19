provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "secrets" {
  source = "../../modules/secrets"
  sgt_node_secret = "${var.sgt_node_secret}"
  sgt_app_secret = "${var.sgt_app_secret}"
  aws_profile = "${var.aws_profile}"
}