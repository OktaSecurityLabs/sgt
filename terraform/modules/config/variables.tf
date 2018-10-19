

variable "aws_profile" {}

variable "aws_region" {
  default = "us-east-1"
}

variable "full_cert_chain" {
  description = "name of the .pem file containing your full ssl cert chain.  should be located in the certs folder"
}

variable "priv_key" {
  description = "name of private key .pem file for ssl certs.  Should be located in certs folder"
}
variable "terraform_backend_bucket_name" {}

variable "environment" {}