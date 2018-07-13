variable "aws_profile" {
}

variable "region" {
  default = "us-east-1"
}

variable "osquery_s3_bucket_name" {

}

variable "client_table_read_capacity" {
  default = 20
}

variable "client_table_write_capacity" {
  default = 20
}

variable "configurations_table_read_capacity" {
  default = 20
}

variable "configurations_table_write_capacity" {
  default = 20
}

variable "distributed_table_read_capacity" {
  default = 20
}

variable "distributed_table_write_capacity" {
  default = 20
}

variable "packqueries_table_read_capacity" {
  default = 20
}

variable "packqueries_table_write_capacity" {
  default = 20
}

variable "querypacks_table_write_capacity" {
  default = 20
}

variable "querypacks_table_read_capacity" {
  default = 20
}

variable "users_table_write_capacity" {
  default = 20
}

variable "users_table_read_capacity" {
  default = 20
}

