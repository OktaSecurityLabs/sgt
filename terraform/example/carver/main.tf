provider "aws" {
  region = "${var.aws_region}"
  profile = "${var.aws_profile}"
}

module "carver" {
  source = "../../modules/carver"
  aws_profile = "${var.aws_profile}"
  aws_region = "${var.aws_region}"
  terraform_backend_bucket_name = "${var.terraform_backend_bucket_name}"
  environment = "${var.environment}"
}