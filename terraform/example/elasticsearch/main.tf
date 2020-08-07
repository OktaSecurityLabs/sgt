provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
  version = ">= 1.21.0"
}

module "elasticsearch" {
  source = "../../modules/elasticsearch"
  user_ip_address = "${var.user_ip_address}"
  aws_profile = "${var.aws_profile}"
  create_elasticsearch = "${var.create_elasticsearch}"
  elasticsearch_instance_count = "${var.elasticsearch_instance_count}"
  elasticsearch_instance_type = "${var.elasticsearch_instance_type}"
  elasticsearch_dedicated_master_count = "${var.elasticsearch_dedicated_master_count}"
  elasticsearch_master_instance_type = "${var.elasticsearch_master_instance_type}"
  elasticsearch_volume_size = "${var.elasticsearch_volume_size}"
  elasticsearch_volume_type = "${var.elasticsearch_volume_type}"
}
