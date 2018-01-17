module "vpc" {
  source = "../../modules/vpc"
  aws_profile = "${var.aws_profile}"
}