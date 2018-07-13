module "datastore" {
  source = "../../modules/datastore"
  aws_profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  osquery_s3_bucket_name = "${var.sgt_osquery_results_bucket_name}"
  client_table_read_capacity = "${var.client_table_read_capacity}"
  client_table_write_capacity = "${var.client_table_write_capacity}"
  configurations_table_read_capacity = "${var.configurations_table_read_capacity}"
  configurations_table_write_capacity = "${var.configurations_table_write_capacity}"
  distributed_table_read_capacity = "${var.distributed_table_read_capacity}"
  distributed_table_write_capacity = "${var.distributed_table_write_capacity}"
  packqueries_table_read_capacity = "${var.distributed_table_read_capacity}"
  packqueries_table_write_capacity = "${var.distributed_table_write_capacity}"
  querypacks_table_read_capacity = "${var.querypacks_table_read_capacity}"
  querypacks_table_write_capacity = "${var.querypacks_table_write_capacity}"
  users_table_read_capacity = "${var.users_table_read_capacity}"
  users_table_write_capacity = "${var.users_table_write_capacity}"
}

