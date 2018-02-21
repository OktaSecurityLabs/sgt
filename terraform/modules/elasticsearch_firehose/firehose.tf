provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
}

data "terraform_remote_state" "elasticsearch" {
  backend = "local"
  config {
    path = "../elasticsearch/terraform.tfstate"
  }
}


resource "aws_s3_bucket" "sgt-osquery_results-s3" {
  bucket = "${var.sgt-s3-osquery-results-bucket-name}"
}

data "aws_iam_policy_document" "sgt-s3_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:ListBucketMultipartUploads",
      "s3:ListBucket",
      "s3:GetObject",
      "s3:GetBucketLocation",
      "s3:AbortMultipartUpload"
    ]
    resources = [
      "${aws_s3_bucket.sgt-osquery_results-s3.arn}/*",
      "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    ]
  }
}

resource "aws_iam_policy" "sgt-firehose-s3-policy" {
  policy = "${data.aws_iam_policy_document.sgt-s3_policy_doc.json}"
  name = "sgt-s3-policy"
}

data "aws_iam_policy_document" "sgt_firehose_assume_role_policy_doc" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }
  }
}


resource "aws_iam_role" "sgt-firehose-assume-role" {
  name = "sgt_firehose_role"
  assume_role_policy = "${data.aws_iam_policy_document.sgt_firehose_assume_role_policy_doc.json}"
}

resource "aws_iam_role_policy_attachment" "attach_s3_policy" {
  policy_arn = "${aws_iam_policy.sgt-firehose-s3-policy.arn}"
  role = "${aws_iam_role.sgt-firehose-assume-role.id}"
}

data "aws_iam_policy_document" "lambda_assume_role_policy_doc"{
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}


data "aws_iam_policy_document" "lambda_policy" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }
}

resource "aws_iam_policy" "lambda_policy" {
  policy = "${data.aws_iam_policy_document.lambda_policy.json}"
}

resource "aws_iam_role" "sgt-osquery-firehose-lambda_role" {
  name = "sgt_firehose_lambda_role"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role_policy_doc.json}"
}

resource "aws_iam_role_policy_attachment" "sgt_lambda_policy_attachment" {
  policy_arn = "${aws_iam_policy.lambda_policy.arn}"
  role = "${aws_iam_role.sgt-osquery-firehose-lambda_role.name}"
}

resource "aws_lambda_function" "sgt_osquery_results_date_transform" {
  function_name = "sgt_osquery_results_date_transform"
  filename = "lambda.zip"
  handler = "main"
  role = "${aws_iam_role.sgt-osquery-firehose-lambda_role.arn}"
  runtime = "go1.x"
  timeout = 120
  source_code_hash = "${base64sha256(file("lambda.zip"))}"
}


resource "aws_kinesis_firehose_delivery_stream" "sgt-firehose-osquery_results" {
  name = "sgt-firehose-osquery_results"
  destination = "elasticsearch"
  #commented out until terraform supports data transformation outside of extended s3.  For the time being, this needs to be enabled via console
  /*extended_s3_configuration {
    role_arn = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    bucket_arn = "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    buffer_size = 5
    buffer_interval = 300
    prefix = "osquery_results"
    processing_configuration {
      enabled = true
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.sgt_osquery_results_date_transform.arn}:$LATEST"
        }
      }
    }
  }*/
  elasticsearch_configuration {
    domain_arn = "${data.terraform_remote_state.elasticsearch.elasticsearch_domain_arn}"
    role_arn = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    index_name = "osquery_results"
    type_name = "osquery_results"
    index_rotation_period = "OneMonth"
    s3_backup_mode = "AllDocuments"
  }

  s3_configuration {
    role_arn = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    bucket_arn = "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    buffer_size = 5
    buffer_interval = 300
    prefix = "osquery_results"
  }

}

resource "aws_kinesis_firehose_delivery_stream" "sgt-firehose-distributed-osquery_results" {
  name = "sgt-firehose-distributed_osquery_results"
  destination = "elasticsearch"

  s3_configuration {
    role_arn = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    bucket_arn = "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    buffer_size = 5
    buffer_interval = 60
    prefix = "distributed_osquery_results"
  }

  elasticsearch_configuration {
    domain_arn = "${data.terraform_remote_state.elasticsearch.elasticsearch_domain_arn}"
    role_arn = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    index_name = "distributed_osquery_results"
    type_name = "osquery_results"
    index_rotation_period = "OneMonth"
    s3_backup_mode = "AllDocuments"
  }
}

## create iam user to allow nodes to send directly to firehose

data "aws_iam_policy_document" "sgt-node-user" {
  statement {
    effect = "Allow"
    actions = [
      "firehose:UpdateDestination",
      "firehose:PutRecord",
      "firehose:PutRecordBatch"
    ],
    resources = [
      "${aws_kinesis_firehose_delivery_stream.sgt-firehose-osquery_results.arn}"
    ]
  }
}

data "aws_iam_policy_document" "elasticsearch_policy" {
  statement {
    effect = "Allow"
    actions = [
      "es:DescribeElasticsearchDomain",
      "es:DescribeElasticsearchDomains",
      "es:DescribeElasticsearchDomainConfig",
      "es:ESHttpPost",
      "es:ESHttpPut"
    ]
    resources = [
      "${data.terraform_remote_state.elasticsearch.elasticsearch_domain_arn}",
      "${data.terraform_remote_state.elasticsearch.elasticsearch_domain_arn}/*"
    ]
  }
}

resource "aws_iam_policy" "elasticsearch_policy" {
  policy = "${data.aws_iam_policy_document.elasticsearch_policy.json}"
}

resource "aws_iam_role_policy_attachment" "elasticsearch_policy_attachment" {
  policy_arn = "${aws_iam_policy.elasticsearch_policy.arn}"
  role = "${aws_iam_role.sgt-firehose-assume-role.name}"
}

resource "aws_iam_policy" "sgt-node-user-policy" {
  policy = "${data.aws_iam_policy_document.sgt-node-user.json}"
}

resource "aws_iam_user" "sgt-node-firehose-user" {
  name = "sgt_node_firehose_user"
}

resource "aws_iam_user_policy_attachment" "attach-node-firehose-policy" {
  policy_arn = "${aws_iam_policy.sgt-node-user-policy.arn}"
  user = "${aws_iam_user.sgt-node-firehose-user.name}"
}

resource "aws_iam_access_key" "node-firehose-access-key" {
  user = "${aws_iam_user.sgt-node-firehose-user.name}"
}
