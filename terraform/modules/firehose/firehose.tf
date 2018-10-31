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
  name = "sgt_firehose_assume_role"
  assume_role_policy = "${data.aws_iam_policy_document.sgt_firehose_assume_role_policy_doc.json}"
}

data "aws_iam_policy_document" "firehose_invoke_lambda_policy_doc" {
  statement {
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction",
      "lambda:GetFunctionConfiguration",
      "logs:PutLogEvents"
    ]
    resources = [
      "${aws_lambda_function.sgt_osquery_results_date_transform.arn}:$LATEST"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "logs:PutLogEvents"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "kinesis:DescribeStream",
      "kinesis:GetShardIterator",
      "kinesis:GetRecords"
    ]
    resources = [
      "${aws_kinesis_firehose_delivery_stream.sgt-firehose-osquery_results.arn}"
    ]
  }
}

resource "aws_iam_policy" "firehose_invoke_lambda_policy" {
  name = "sgt-firehose-lambda-policy"
  policy = "${data.aws_iam_policy_document.firehose_invoke_lambda_policy_doc.json}"
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
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    bucket_arn = "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    buffer_interval = 60
    buffer_size = 10
    prefix = "osquery_results"
    processing_configuration = [
      {
        enabled = "true"
        processors = [
          {
            type = "Lambda"
            parameters = [
              {
                parameter_name = "LambdaArn"
                parameter_value = "${aws_lambda_function.sgt_osquery_results_date_transform.arn}:$LATEST"
              }
            ]
          }
        ]
      }
    ]
  }
}

resource "aws_kinesis_firehose_delivery_stream" "sgt-firehose-distributed-osquery_results" {
  name = "sgt-firehose-distributed_osquery_results"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = "${aws_iam_role.sgt-firehose-assume-role.arn}"
    bucket_arn = "${aws_s3_bucket.sgt-osquery_results-s3.arn}"
    buffer_interval = 60
    buffer_size = 10
    prefix = "distributed_osquery_results"
    processing_configuration = [
      {
        enabled = "true"
        processors = [
          {
            type = "Lambda"
            parameters = [
              {
                parameter_name = "LambdaArn"
                parameter_value = "${aws_lambda_function.sgt_osquery_results_date_transform.arn}:$LATEST"
              }
            ]
          }
        ]
      }
    ]
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

resource "aws_iam_role_policy_attachment" "firehose_invoke_lambda_policy_attachment" {
  policy_arn = "${aws_iam_policy.firehose_invoke_lambda_policy.arn}"
  role = "${aws_iam_role.sgt-firehose-assume-role.name}"
}

resource "aws_iam_policy" "sgt-node-user-policy" {
  name = "sgt-node-user-policy"
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
