output "dynamo_table_osquery_clients_arn" {
  value = "${aws_dynamodb_table.clients.arn}"
}

output "dynamo_table_osquery_configurations_arn" {
  value = "${aws_dynamodb_table.osquery_configurations.arn}"
}

output "dynamo_table_osquery_distributed_queries_arn" {
  value = "${aws_dynamodb_table.osquery_distributed_queries.arn}"
}

output "dynamo_table_osquery_packqueries_arn" {
  value = "${aws_dynamodb_table.osquery_packqueries.arn}"
}

output "dynamo_table_osquery_querypacks_arn" {
  value = "${aws_dynamodb_table.osquery_querypacks.arn}"
}

output "dynamo_table_osquery_users_arn" {
  value = "${aws_dynamodb_table.osquery_users.arn}"
}

output "s3_bucket_arn" {
  value = "${aws_s3_bucket.osquery_s3_bucket.arn}"
}

output "s3_bucket_name" {
  value = "${aws_s3_bucket.osquery_s3_bucket.bucket}"
}