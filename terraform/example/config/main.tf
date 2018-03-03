module "config" {
  source = "../../modules/config"
  aws_profile = "${var.aws_profile}"
  full_cert_chain = "${var.full_ssl_certchain}"
  priv_key = "${var.ssl_private_key}"
}