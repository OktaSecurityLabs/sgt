module "datastore" {
  source = "../../modules/datastore"
  aws_profile = "${var.aws_profile}"
  region = "${var.aws_region}"
}

