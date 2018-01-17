module "secrets" {
  source = "../../modules/secrets"
  sgt_node_secret = "${var.sgt_node_secret}"
  sgt_app_secret = "${var.sgt_app_secret}"
  aws_profile = "${var.aws_profile}"
}