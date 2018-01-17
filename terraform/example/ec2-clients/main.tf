module "ec2" {
  source = "../../modules/ec2-clients"
  num_clients = "${var.num_clients}"
}
