provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

data "terraform_remote_state" "firehose" {
  backend = "s3"
  config {
    bucket = "${var.terraform_backend_bucket_name}"
    key = "${var.environment}/elasticsearch_firehose/terraform.tfstate"
    profile = "${var.aws_profile}"
    region = "${var.aws_region}"
  }
}

data "terraform_remote_state" "datastore" {
  backend = "s3"
  config {
    bucket = "${var.terraform_backend_bucket_name}"
    key = "${var.environment}/datastore/terraform.tfstate"
    profile = "${var.aws_profile}"
    region = "${var.aws_region}"
  }
}

resource "aws_s3_bucket_object" "osquery-sgt-binary" {
  bucket = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  source = "../../../sgt"
  key = "sgt/sgt"
  etag = "${md5(file("../../../sgt"))}"
}

data "template_file" "sgt-config-file" {
  template = "${file("${path.module}/example.config.json")}"
  vars {
    firehose_aws_access_key_id = "${data.terraform_remote_state.firehose.sgt-node-user-access-key-id}"
    firehose_aws_secret_access_key = "${data.terraform_remote_state.firehose.sgt-node-user-secret-access-key}",
    firehose_stream_name = "${data.terraform_remote_state.firehose.sgt-firehose-stream-name}",
    distributed_query_logger_firehose_stream_name = "${data.terraform_remote_state.firehose.sgt-distributed-firehose-stream-name}"
    auto_approve_nodes = "${var.auto_approve_nodes}"
  }
}

resource "aws_s3_bucket_object" "osquery-sgt-config" {
  bucket = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  content = "${data.template_file.sgt-config-file.rendered}"
  key = "sgt/config.json"
  etag = "${md5("{data.template_file.sgt-config-file.rendered}")}"
}

resource "aws_s3_bucket_object" "osquery-sgt-fullchain_pem" {
  bucket = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  source = "../../../certs/${var.full_cert_chain}"
  key = "sgt/fullchain.pem"
  etag = "${md5(file("../../../certs/${var.full_cert_chain}"))}"
}

resource "aws_s3_bucket_object" "osquery-sgt-privkey_pem" {
  bucket = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  source = "../../../certs/${var.priv_key}"
  key = "sgt/privkey.pem"
  etag = "${md5(file("../../../certs/${var.priv_key}"))}"
}
