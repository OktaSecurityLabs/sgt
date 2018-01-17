provider "aws" {
  profile = "${var.aws_profile}"
  region = "us-east-1"
}

resource "aws_dynamodb_table" "clients" {
  name = "osquery_clients"
  read_capacity = 20
  write_capacity = 20
  hash_key = "node_key"

  attribute {
    name = "node_key"
    type = "S"
  }
}



resource "aws_dynamodb_table" "osquery_configurations" {
  name = "osquery_configurations"
  hash_key = "config_name"
  read_capacity = 20
  write_capacity = 20

  attribute {
    name = "config_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_distributed_queries" {
  name = "osquery_distributed_queries"
  hash_key = "node_key"
  read_capacity = 20
  write_capacity = 20

  attribute {
    name = "node_key"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_packqueries" {
  name = "osquery_packqueries"
  hash_key = "query_name"
  write_capacity = 20
  read_capacity = 20

  attribute {
    name = "query_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_querypacks" {
  name = "osquery_querypacks"
  hash_key = "pack_name"
  write_capacity = 20
  read_capacity = 20

  attribute {
    name = "pack_name"
    type = "S"
  }
}

resource "aws_dynamodb_table" "osquery_users" {
  name = "osquery_users"
  hash_key = "username"
  write_capacity = 5
  read_capacity = 5

  attribute {
    name = "username"
    type = "S"
  }
}