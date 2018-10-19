provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "config" {
  source = "../../modules/elasticsearch_config"
  aws_profile = "${var.aws_profile}"
  full_cert_chain = "${var.full_ssl_certchain}"
  priv_key = "${var.ssl_private_key}"
  terraform_backend_bucket_name = "${var.terraform_backend_bucket_name}"
  environment = "${var.environment}"
}