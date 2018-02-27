module "s3" {
  source = "../../modules/elasticsearch_s3"
  osquery_s3_bucket_name = "${var.sgt_config_bucket_name}"
  aws_profile = "${var.aws_profile}"
  full_cert_chain = "${var.full_ssl_certchain}"
  priv_key = "${var.ssl_private_key}"
}