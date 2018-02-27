module "datastore" {
  source = "../../modules/datastore"
  aws_profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  osquery_s3_bucket_name = "${var.sgt_osquery_results_bucket_name}"
}

