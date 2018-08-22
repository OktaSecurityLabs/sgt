provider "aws" {
  region = "us-east-1"
  profile = "${var.aws_profile}"
  version = ">= 1.21.0"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {
  current = true
}

data "terraform_remote_state" "vpc" {
  backend = "local"
  config {
    path = "../vpc/terraform.tfstate"
  }
}

data "terraform_remote_state" "datastore" {
  backend = "local"
  config {
    path = "../datastore/terraform.tfstate"
  }
}

data "terraform_remote_state" "s3" {
  backend = "local"
  config {
    path = "../config/terraform.tfstate"
  }
}

data "terraform_remote_state" "ssm" {
  backend = "local"
  config {
    path = "../secrets/terraform.tfstate"
  }
}

data "terraform_remote_state" "firehose" {
  backend = "local"
  config {
    path = "../firehose/terraform.tfstate"
  }
}

data "aws_iam_policy_document" "server_assume_role_policy_doc" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "server_dynamo_access_policy_doc" {
  statement {
    actions = [
      "dynamodb:*",
      "dax:*"
    ]
    effect = "Allow"
    resources = [
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_clients_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_configurations_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_distributed_queries_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_packqueries_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_querypacks_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_osquery_users_arn}"
    ]
  }
}

data "aws_iam_policy_document" "osquery_s3_role_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "s3:AbortMultipartUpload",
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:ListBucketMultipartUploads",
      "s3:PutObject"
    ]
    resources = [
      "${data.terraform_remote_state.datastore.s3_bucket_arn}",
      "${data.terraform_remote_state.datastore.s3_bucket_arn}/*"
    ]
  }
}

data "aws_iam_policy_document" "osquery_ssm_param_store_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:*",
      "logs:*"
    ]
    #resources = ["*"]
    resources = [
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/${data.terraform_remote_state.ssm.sgt_node_secret_id}",
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/${data.terraform_remote_state.ssm.sgt_app_secret_id}"
    ]
  }
}


data "aws_iam_policy_document" "osquery_firehose_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "firehose:UpdateDestination",
      "firehose:PutRecordBatch",
      "firehose:PutRecord"
    ]
    resources = [
      "${data.terraform_remote_state.firehose.sgt-distributed-firehose-stream-arn}"
    ]
  }
}

resource "aws_iam_policy" "osquery_s3_policy" {
  name = "osquery-sgt-s3-access-policy"
  policy = "${data.aws_iam_policy_document.osquery_s3_role_policy_doc.json}"
}

resource "aws_iam_policy" "server_dynamo_access_policy" {
  name = "osquery_sgt_dynamo_access"
  policy = "${data.aws_iam_policy_document.server_dynamo_access_policy_doc.json}"
}

resource "aws_iam_role" "server_assume_role" {
  name = "osquery_sgt_assume_role"
  assume_role_policy = "${data.aws_iam_policy_document.server_assume_role_policy_doc.json}"
}

resource "aws_iam_policy" "osquery_ssm_param_store_policy" {
  name = "osquery_sgt_ssm_policy"
  policy = "${data.aws_iam_policy_document.osquery_ssm_param_store_policy_doc.json}"
}

resource "aws_iam_policy" "osquery_firehose_policy" {
  name = "osquery_sgt_firehose_policy"
  policy = "${data.aws_iam_policy_document.osquery_firehose_policy_doc.json}"
}

resource "aws_iam_role_policy_attachment" "attach_policy" {
  role = "${aws_iam_role.server_assume_role.name}"
  policy_arn = "${aws_iam_policy.server_dynamo_access_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "attach_s3_policy" {
  role = "${aws_iam_role.server_assume_role.name}"
  policy_arn = "${aws_iam_policy.osquery_s3_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "attach_ssm_policy" {
  role = "${aws_iam_role.server_assume_role.name}"
  policy_arn = "${aws_iam_policy.osquery_ssm_param_store_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "attach_firehose_policy" {
  role ="${aws_iam_role.server_assume_role.name}"
  policy_arn = "${aws_iam_policy.osquery_firehose_policy.arn}"
}

resource "aws_iam_instance_profile" "osquery_sgt_instance_profile" {
  name = "osquery_sgt_instance_profile"
  role = "${aws_iam_role.server_assume_role.name}"
}
resource "aws_subnet" "WebappALBSubnet_us_east_1a" {
  vpc_id = "${data.terraform_remote_state.vpc.vpc_id}"
  cidr_block = "10.0.20.0/24"
  map_public_ip_on_launch = false
  availability_zone = "us-east-1a"
  tags {
    Name = "sgt-ELB_us_east_1a"
  }
}

resource "aws_subnet" "WebappALBSubnet_us_east_1b" {
  vpc_id = "${data.terraform_remote_state.vpc.vpc_id}"
  cidr_block = "10.0.21.0/24"
  map_public_ip_on_launch = false
  availability_zone = "us-east-1b"
  tags {
    Name = "sgt-ELB_us_east_1b"
  }
}

resource "aws_security_group" "elb_security_group" {
  name = "sgt_elb_security_group"
  vpc_id = "${data.terraform_remote_state.vpc.vpc_id}"
  ingress {
    from_port = 443
    protocol = "tcp"
    to_port = 443
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port = 0
    protocol = "-1"
    to_port = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "external_to_elb" {
  name = "sgt-elb-ingress"
  vpc_id = "${data.terraform_remote_state.vpc.vpc_id}"
}
resource "aws_security_group_rule" "external_to_elb_ingress" {
  security_group_id = "${aws_security_group.external_to_elb.id}"
  type = "ingress"
  protocol = "TCP"
  from_port = 443
  to_port = 443
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "external_to_elb_egress" {
  security_group_id = "${aws_security_group.external_to_elb.id}"
  type = "egress"
  protocol = "-1"
  from_port = 0
  to_port = 0
  cidr_blocks = ["0.0.0.0/0"]
}
resource "aws_elb" "osquery-sgt_elb" {
  name = "osquery-sgt-loadblancer"
  security_groups = [
    "${aws_security_group.external_to_elb.id}",
    "${aws_security_group.elb_security_group.id}"
  ]
  subnets = [
    "${data.terraform_remote_state.vpc.sgt_public_subnet_us_east_1a_id}",
    "${data.terraform_remote_state.vpc.sgt_public_subnet_us_east_1b_id}"
  ]
  tags {
    Name = "osquery-sgt-elb"
  }
  listener {
    instance_port = 443
    instance_protocol = "TCP"
    lb_port = 443
    lb_protocol = "TCP"
  }
  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 4
    timeout = 3
    target = "TCP:443"
    interval = 30
  }
}


resource "aws_security_group" "load_balancer_to_sgt_sg" {
  name = "sgt_permit_load_balancer"
  vpc_id = "${data.terraform_remote_state.vpc.vpc_id}"
}

resource "aws_security_group_rule" "load_balancer_to_sgt_ingress" {
  security_group_id = "${aws_security_group.load_balancer_to_sgt_sg.id}"
  type = "ingress"
  protocol = "TCP"
  from_port = 443
  to_port = 443
  source_security_group_id = "${aws_security_group.external_to_elb.id}"
}


resource "aws_security_group_rule" "load_balancer_to_sg_egress" {
  security_group_id = "${aws_security_group.load_balancer_to_sgt_sg.id}"
  type = "egress"
  protocol = "-1"
  from_port = 0
  to_port = 0
  cidr_blocks = ["0.0.0.0/0"]
}

data "template_file" "user_data" {
  template = "${file("userdata.sh")}"
  vars {
    bucket_name = "${data.terraform_remote_state.datastore.s3_bucket_name}"
  }
}

resource "aws_launch_configuration" "osquery_sgt_LaunchConfig" {
  image_id        = "${var.asg_ami_id}"
  iam_instance_profile = "${aws_iam_instance_profile.osquery_sgt_instance_profile.arn}"
  instance_type   = "${var.instance_type}"
  security_groups = ["${aws_security_group.load_balancer_to_sgt_sg.id}"]
  user_data       = "${data.template_file.user_data.rendered}"
  key_name = "${var.instance_ssh_key_name}"
  lifecycle {
      create_before_destroy = true
    }
  }

resource "aws_autoscaling_group" "osquery-sgt_asg" {
  name = "sgt-${aws_launch_configuration.osquery_sgt_LaunchConfig.name}"
  vpc_zone_identifier = [
    "${data.terraform_remote_state.vpc.sgt_private_us_east_1a_id}",
    "${data.terraform_remote_state.vpc.sgt_private_us_east_1b_id}"
  ]
  #name_prefix = "osquery-sgt-1-"
  lifecycle {
    create_before_destroy = true
  }
  load_balancers = ["${aws_elb.osquery-sgt_elb.name}"]
  min_size = "${var.asg_min_size}"
  max_size = "${var.asg_max_size}"
  wait_for_elb_capacity = 2
  health_check_grace_period = 300
  health_check_type = "ELB"
  desired_capacity = "${var.asg_desired_size}"
  #placement_group = "${aws_placement_group.osquery-sgt_placement_group.id}"
  launch_configuration = "${aws_launch_configuration.osquery_sgt_LaunchConfig.name}"
  tag {
    key                 = "Name"
    value               = "sgt-${aws_launch_configuration.osquery_sgt_LaunchConfig.name}"
    propagate_at_launch = true
  }
}

data "aws_route53_zone" "osquery-sgt-dns-zone" {
  name = "${var.dns_zone_domain}"
}

resource "aws_route53_record" "osquery-sgt-subdomain" {
  name = "${var.dns_subdomain}"
  zone_id = "${data.aws_route53_zone.osquery-sgt-dns-zone.id}"
  type = "A"
  alias {
    name = "${aws_elb.osquery-sgt_elb.dns_name}"
    zone_id = "${aws_elb.osquery-sgt_elb.zone_id}"
    evaluate_target_health = true
  }
}
