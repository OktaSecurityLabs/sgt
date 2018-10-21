variable "instance_ssh_key_name" {
  description = "ssh key for access to ec2 instances"
}

variable "instance_type" {
  description = "type of ec2 instance to use"
}
variable "asg_ami_id" {
  default = "ami-cd0f5cb6"
  description = "ami id to launch asg with (must be Ubuntu 14.04 or later)"
}
variable "asg_min_size" {
  description = "minimum number of autoscaling group instances to have at a given time"
}
variable "asg_max_size" {
  description = "maximum number of autoscaling group instances to have at a given time"
}

variable "asg_desired_size" {
  default = 2
}

variable "alb_private_subnet_cidr_us_east_1a" {
  description = "subnet for use by alb and ec2 instance"
  default = "10.11.12.0/24"
}

variable "abl_private_subnet_cidr_us_east_1b" {
  description = "subnet for use by alb and ec2 instance"
  default = "10.11.13.0/24"
}

variable "elb_us_east_1a_public_subnet" {}

variable "elb_us_east_1b_public_subnet" {}

variable "dns_zone_domain" {
  description = "dns zone to use for dns records"
}

variable "dns_subdomain" {
  description = "subdomain to be used for server"
}


variable "aws_profile" {}
variable "aws_region" {}
variable "terraform_backend_bucket_name" {}

variable "environment" {}
#variable "nat_gw_id" {}
