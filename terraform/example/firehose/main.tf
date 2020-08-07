provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "firehose" {
  source = "../../modules/firehose"
  sgt-s3-osquery-results-bucket-name = "${var.sgt_osquery_results_bucket_name}"
  aws_profile = "${var.aws_profile}"
  create_elasticsearch = "${var.create_elasticsearch}"
  terraform_backend_bucket_name = "${var.terraform_backend_bucket_name}"
  environment = "${var.environment}"
}