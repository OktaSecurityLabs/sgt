provider "aws" {
  region = "${var.region}"
  profile = "${var.aws_profile}"
}

resource "aws_ssm_parameter" "sgt_node_secret" {
  name = "sgt_node_secret"
  type = "SecureString"
  value = "${var.sgt_node_secret}"
  overwrite = true
}

resource "aws_ssm_parameter" "sgt_app_secret" {
  name = "sgt_app_secret"
  type = "SecureString"
  value = "${var.sgt_app_secret}"
  overwrite = true
}