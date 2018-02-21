module "firehose" {
  source = "../../modules/firehose"
  sgt-s3-osquery-results-bucket-name = "${var.sgt_osquery_results_bucket_name}"
  aws_profile = "${var.aws_profile}"
  create_elasticsearch = "${var.create_elasticsearch}"
}