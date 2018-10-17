output "dynamo_table_osquery_clients_arn" {
  value = "${module.datastore.dynamo_table_osquery_clients_arn}"
}

output "dynamo_table_osquery_configurations_arn" {
  value = "${module.datastore.dynamo_table_osquery_configurations_arn}"
}

output "dynamo_table_osquery_distributed_queries_arn" {
  value = "${module.datastore.dynamo_table_osquery_distributed_queries_arn}"
}

output "dynamo_table_osquery_packqueries_arn" {
  value = "${module.datastore.dynamo_table_osquery_packqueries_arn}"
}

output "dynamo_table_osquery_querypacks_arn" {
  value = "${module.datastore.dynamo_table_osquery_querypacks_arn}"
}

output "dynamo_table_osquery_users_arn" {
  value = "${module.datastore.dynamo_table_osquery_users_arn}"
}

output "dynamo_table_filecarves_arn" {
  value = "${module.datastore.dynamo_table_filecarves_arn}"
}

output "dynamo_table_carve_data_arn" {
  value = "${module.datastore.dynamo_table_carve_data_arn}"
}

output "s3_bucket_arn" {
  value = "${module.datastore.s3_bucket_arn}"
}

output "s3_bucket_name" {
  value = "${module.datastore.s3_bucket_name}"
}