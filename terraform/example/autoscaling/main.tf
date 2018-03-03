module "autoscaling" {
  source = "../../modules/autoscaling"
  instance_ssh_key_name = "${var.aws_keypair}"
  asg_min_size = "${var.asg_min_size}"
  asg_max_size = "${var.asg_max_size}"
  asg_desired_size = "${var.asg_desired_size}"
  instance_type = "t2.micro"
  #make sure you set a subnet that will work in your VPC
  alb_private_subnet_cidr_us_east_1a = "10.0.10.0/24"
  abl_private_subnet_cidr_us_east_1b = "10.0.11.0/24"
  elb_us_east_1a_public_subnet = "10.0.12.0/24"
  elb_us_east_1b_public_subnet = "10.0.13.0/24"
  dns_zone_domain = "${var.domain}"
  dns_subdomain = "${var.subdomain}"
  aws_profile = "${var.aws_profile}"
}

