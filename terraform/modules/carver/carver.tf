data "terraform_remote_state" "datastore" {
  backend = "s3"
  config {
    bucket = "${var.terraform_backend_bucket_name}"
    key = "${var.environment}/datastore/terraform.tfstate"
    region = "${var.aws_region}"
    profile = "${var.aws_profile}"
  }
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

data "aws_iam_policy_document" "carve_builder_lambda_policy" {
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
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:ListBucket"
    ]
    resources = [
      "${data.terraform_remote_state.datastore.s3_bucket_arn}/*"
    ]
  }
  statement {
    actions = [
      "dynamodb:*",
      "dax:*"
    ]
    effect = "Allow"
    resources = [
      "${data.terraform_remote_state.datastore.dynamo_table_filecarves_arn}",
      "${data.terraform_remote_state.datastore.dynamo_table_carve_data_arn}"
    ]
  }
}

resource "aws_iam_policy" "carve_builder_lambda_policy" {
  name = "carve-builder-lambda-policy"
  policy = "${data.aws_iam_policy_document.carve_builder_lambda_policy.json}"
}

resource "aws_iam_role" "carve_builder_lambda_iam_role" {
  name = "carve-builder-lambda-role"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role_policy_doc.json}"
}

resource "aws_iam_role_policy_attachment" "carve_builder_lambda_policy_attachment" {
  policy_arn = "${aws_iam_policy.carve_builder_lambda_policy.arn}"
  role = "${aws_iam_role.carve_builder_lambda_iam_role.name}"
}


data "archive_file" "carve_builder_lambda_zip" {
  type = "zip"
  source_file = "../../../lambda_functions/carvebuilder/main"
  output_path = "../../../lambda_functions/carvebuilder/deployment.zip"
}


resource "aws_lambda_function" "carve_builder_lambda_function" {
  depends_on = [
    "data.archive_file.carve_builder_lambda_zip"
  ]
  function_name = "carve-builder-lambda"
  handler = "main"
  role = "${aws_iam_role.carve_builder_lambda_iam_role.arn}"
  runtime = "go1.x"
  filename = "${data.archive_file.carve_builder_lambda_zip.output_path}"
  source_code_hash = "${data.archive_file.carve_builder_lambda_zip.output_base64sha256}"
  environment {
    variables {
      CARVE_BUCKET = "${data.terraform_remote_state.datastore.s3_bucket_name}"
    }
  }
  timeout = 300
  memory_size = 3008
}


data "aws_iam_policy_document" "carve_manager_lambda_policy" {
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
  statement {
    actions = [
      "dynamodb:*",
      "dax:*"
    ]
    effect = "Allow"
    resources = [
      "${data.terraform_remote_state.datastore.dynamo_table_filecarves_arn}",
    ]
  }
  statement {
    actions = [
      "lambda:InvokeFunction"
    ]
    effect = "Allow"
    resources = [
      "${aws_lambda_function.carve_builder_lambda_function.arn}"
    ]
  }
}

resource "aws_iam_policy" "carve_manager_lambda_policy" {
  name = "carve-manager-lambda-policy"
  policy = "${data.aws_iam_policy_document.carve_manager_lambda_policy.json}"
}

resource "aws_iam_role" "carve_manager_lambda_iam_role" {
  name = "carve-manager-lambda-role"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role_policy_doc.json}"
}

resource "aws_iam_role_policy_attachment" "carve_manager_lambda_policy_attachment" {
  policy_arn = "${aws_iam_policy.carve_manager_lambda_policy.arn}"
  role = "${aws_iam_role.carve_manager_lambda_iam_role.name}"
}

data "archive_file" "carve_manager_lambda_zip" {
  type = "zip"
  source_file = "../../../lambda_functions/carvemanager/main"
  output_path = "../../../lambda_functions/carvemanager/deployment.zip"
}

resource "aws_lambda_function" "carve_manager_lambda_function" {
  depends_on = [
    "data.archive_file.carve_manager_lambda_zip"
  ]
  function_name = "carve-manager-lambda"
  handler = "main"
  role = "${aws_iam_role.carve_manager_lambda_iam_role.arn}"
  runtime = "go1.x"
  filename = "${data.archive_file.carve_manager_lambda_zip.output_path}"
  source_code_hash = "${data.archive_file.carve_manager_lambda_zip.output_base64sha256}"
  environment {
    variables {
      CARVE_BUILDER = "${aws_lambda_function.carve_builder_lambda_function.function_name}"
    }
  }
  timeout = 300
  memory_size = 256
}

resource "aws_cloudwatch_event_rule" "every_minute_carve_manager" {
    name = "every_minute_carve_manager_invoke"
    description = "runs once a minute"
    schedule_expression = "rate(1 minute)"
}

resource "aws_cloudwatch_event_target" "run_carve_manager_lambda" {
    rule = "${aws_cloudwatch_event_rule.every_minute_carve_manager.name}"
    target_id = "check_foo"
    arn = "${aws_lambda_function.carve_manager_lambda_function.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_carve_manager" {
    statement_id = "AllowExecutionFromCloudWatch"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.carve_manager_lambda_function.arn}"
    principal = "events.amazonaws.com"
    source_arn = "${aws_cloudwatch_event_rule.every_minute_carve_manager.arn}"
}