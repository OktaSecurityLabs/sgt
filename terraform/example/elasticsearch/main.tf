module "elasticsearch" {
  source = "../../modules/elasticsearch"
  user_ip_address = "${var.user_ip_address}"
  aws_profile = "${var.aws_profile}"
}