terraform {
  backend "s3" {
    bucket = "example-backend-bucket-name"
    key = "example-terraform.tfstate"
  }
}