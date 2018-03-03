variable "aws_profile" {
}

variable "aws_keypair" {}

variable "domain" {}

variable "subdomain" {}

variable "asg_min_size"  {
  default = 2
}

variable "asg_max_size" {
  default = 4
}

variable "asg_desired_size" {}