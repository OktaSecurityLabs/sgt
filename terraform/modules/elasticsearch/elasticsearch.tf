provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
}

resource "aws_elasticsearch_domain" "sgt-osquery_results" {
  domain_name = "${var.elasticsearch_domain_name}"
  elasticsearch_version = "5.5"
  cluster_config {
    instance_count = 3
    instance_type = "m4.large.elasticsearch"
    dedicated_master_enabled = true
    dedicated_master_count = 3
    dedicated_master_type = "t2.medium.elasticsearch"
  }
  ebs_options {
    ebs_enabled = true
    volume_size = 200
    volume_type = "gp2"
  }
}

data "aws_iam_policy_document" "es_domain_policy" {
  statement {
    effect = "Allow"
    principals {
      type = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "es:*"
    ]
    resources = [
      "${aws_elasticsearch_domain.sgt-osquery_results.arn}/*"
    ]
    condition {
      test = "IpAddress"
      values = [
        "${var.user_ip_address}",
        "12.97.85.90/29"
      ]
      variable = "aws:SourceIp"
    }
  }
}

resource "aws_elasticsearch_domain_policy" "elasticsearch_domain_policy" {
  domain_name = "${aws_elasticsearch_domain.sgt-osquery_results.domain_name}"
  access_policies = "${data.aws_iam_policy_document.es_domain_policy.json}"
  depends_on = ["aws_elasticsearch_domain.sgt-osquery_results"]
}


