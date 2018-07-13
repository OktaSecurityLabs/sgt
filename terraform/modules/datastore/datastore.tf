provider "aws" {
  profile = "${var.aws_profile}"
  region = "us-east-1"
}

resource "aws_s3_bucket" "osquery_s3_bucket" {
  bucket = "${var.osquery_s3_bucket_name}"
  acl = "private"
}

resource "aws_dynamodb_table" "clients" {
  name = "osquery_clients"
  read_capacity = "${var.client_table_read_capacity}"
  write_capacity = "${var.client_table_write_capacity}"
  hash_key = "node_key"

  attribute {
    name = "node_key"
    type = "S"
  }
}



resource "aws_dynamodb_table" "osquery_configurations" {
  name = "osquery_configurations"
  hash_key = "config_name"
  read_capacity = "${var.configurations_table_read_capacity}"
  write_capacity = "${var.configurations_table_write_capacity}"

  attribute {
    name = "config_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_distributed_queries" {
  name = "osquery_distributed_queries"
  hash_key = "node_key"
  read_capacity = "${var.distributed_table_read_capacity}"
  write_capacity = "${var.distributed_table_write_capacity}"

  attribute {
    name = "node_key"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_packqueries" {
  name = "osquery_packqueries"
  hash_key = "query_name"
  write_capacity = "${var.packqueries_table_write_capacity}"
  read_capacity = "${var.packqueries_table_read_capacity}"

  attribute {
    name = "query_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_querypacks" {
  name = "osquery_querypacks"
  hash_key = "pack_name"
  write_capacity = "${var.querypacks_table_write_capacity}"
  read_capacity = "${var.querypacks_table_read_capacity}"

  attribute {
    name = "pack_name"
    type = "S"
  }
}

resource "aws_dynamodb_table" "osquery_users" {
  name = "osquery_users"
  hash_key = "username"
  write_capacity = "${var.users_table_write_capacity}"
  read_capacity = "${var.users_table_read_capacity}"

  attribute {
    name = "username"
    type = "S"
  }
}