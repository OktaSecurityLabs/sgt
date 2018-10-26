variable "user_ip_address" {}
variable "aws_profile" {}
variable "create_elasticsearch" {}
variable "aws_region" {}

variable "elasticsearch_instance_count" {
  default = 3
}

variable "elasticsearch_instance_type" {
  default = "m4.large.elasticsearch"
}

variable "elasticsearch_dedicated_master_count" {
  default = 3
}

variable "elasticsearch_master_instance_type" {
  default = "t2.medium.elasticsearch"
}

variable "elasticsearch_volume_size" {
  default = 300
}

variable "elasticsearch_volume_type" {
  default = "gp2"
}

### volume limits: https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-limits.html#ebsresource
