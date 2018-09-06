data "terraform_remote_state" "firehose" {
  backend = "local"
  config {
    path = "../firehose/terraform.tfstate"
  }
}

data "terraform_remote_state" "datastore" {
  backend = "local"
  config {
    path = "../datastore/terraform.tfstate"
  }
}

provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
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
  }
}

resource "aws_s3_bucket_object" "osquery-sgt-config" {
  bucket = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  content = "${data.template_file.sgt-config-file.rendered}"
  key = "sgt/config.json"
  etag = "${md5(base64encode("{data.template_file.sgt-config-file.rendered}"))}"
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
