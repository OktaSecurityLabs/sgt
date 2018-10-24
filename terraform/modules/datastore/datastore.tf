resource "aws_s3_bucket" "osquery_s3_bucket" {
  bucket = "${var.osquery_s3_bucket_name}"
  acl = "private"
}

resource "aws_dynamodb_table" "clients" {
  name = "osquery_clients"
  read_capacity = "${var.client_table_read_capacity}"
  write_capacity = "${var.client_table_write_capacity}"
  hash_key = "node_key"

  attribute {
    name = "node_key"
    type = "S"
  }
}



resource "aws_dynamodb_table" "osquery_configurations" {
  name = "osquery_configurations"
  hash_key = "config_name"
  read_capacity = "${var.configurations_table_read_capacity}"
  write_capacity = "${var.configurations_table_write_capacity}"

  attribute {
    name = "config_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_distributed_queries" {
  name = "osquery_distributed_queries"
  hash_key = "node_key"
  read_capacity = "${var.distributed_table_read_capacity}"
  write_capacity = "${var.distributed_table_write_capacity}"

  attribute {
    name = "node_key"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_packqueries" {
  name = "osquery_packqueries"
  hash_key = "query_name"
  write_capacity = "${var.packqueries_table_write_capacity}"
  read_capacity = "${var.packqueries_table_read_capacity}"

  attribute {
    name = "query_name"
    type = "S"
  }
}


resource "aws_dynamodb_table" "osquery_querypacks" {
  name = "osquery_querypacks"
  hash_key = "pack_name"
  write_capacity = "${var.querypacks_table_write_capacity}"
  read_capacity = "${var.querypacks_table_read_capacity}"

  attribute {
    name = "pack_name"
    type = "S"
  }
}

resource "aws_dynamodb_table" "osquery_users" {
  name = "osquery_users"
  hash_key = "username"
  write_capacity = "${var.users_table_write_capacity}"
  read_capacity = "${var.users_table_read_capacity}"

  attribute {
    name = "username"
    type = "S"
  }
}


resource "aws_dynamodb_table" "filecarves" {
  "attribute" {
    name = "session_id"
    type = "S"
  }
  hash_key = "session_id"
  name = "filecarves"
  read_capacity = 40
  write_capacity = 40
}

resource "aws_dynamodb_table" "carve_data" {
  attribute {
    name = "session_block_id"
    type = "S"
  }

  hash_key = "session_block_id"
  name = "carve_data"
  read_capacity = 25
  write_capacity = 25

  ttl {
    attribute_name = "time_to_live"
    enabled = true
  }
}

data "aws_iam_policy_document" "dynamodb_autoscaling_policy_document" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:DescribeLimits",
      "dynamodb:DescribeTable",
      "dynamodb:UpdateTable"
    ]
    resources = [
      "${aws_dynamodb_table.carve_data.arn}",
      "${aws_dynamodb_table.carve_data.arn}/*"
    ]
  }
}

data "aws_iam_policy_document" "dynamodb_autoscaling_assume_role_document" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["dynamodb.amazonaws.com"]
      type = "Service"
    }
  }
}

resource "aws_iam_role" "dynamodb_autoscaling_role" {
  assume_role_policy = "${data.aws_iam_policy_document.dynamodb_autoscaling_assume_role_document.json}"
}

resource "aws_iam_policy" "dynamodb_autoscaling_policy" {
  name = "sgt-dynamodb-carve-data-autoscaling"
  policy = "${data.aws_iam_policy_document.dynamodb_autoscaling_policy_document.json}"
}

resource "aws_iam_role_policy_attachment" "dynamodb_carve_data-autoscaling_attachment" {
  policy_arn = "${aws_iam_policy.dynamodb_autoscaling_policy.arn}"
  role = "${aws_iam_role.dynamodb_autoscaling_role.name}"
}

resource "aws_appautoscaling_target" "dynamodb_carve_data_read_target" {
  max_capacity       = 2000
  min_capacity       = 25
  resource_id        = "table/${aws_dynamodb_table.carve_data.name}"
  role_arn           = "${aws_iam_role.dynamodb_autoscaling_role.arn}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "dynamodb_table_read_policy" {
  name               = "DynamoDBReadCapacityUtilization:${aws_appautoscaling_target.dynamodb_carve_data_read_target.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = "${aws_appautoscaling_target.dynamodb_carve_data_read_target.resource_id}"
  scalable_dimension = "${aws_appautoscaling_target.dynamodb_carve_data_read_target.scalable_dimension}"
  service_namespace  = "${aws_appautoscaling_target.dynamodb_carve_data_read_target.service_namespace}"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBReadCapacityUtilization"
    }

    target_value = 70
  }
}

resource "aws_appautoscaling_target" "dynamodb_carve_data_write_target" {
  max_capacity       = 2000
  min_capacity       = 25
  resource_id        = "table/${aws_dynamodb_table.carve_data.name}"
  role_arn           = "${aws_iam_role.dynamodb_autoscaling_role.arn}"
  scalable_dimension = "dynamodb:table:WriteCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_policy" "dynamodb_carve_data_write_policy" {
  name               = "DynamoDBWriteCapacityUtilization:${aws_appautoscaling_target.dynamodb_carve_data_write_target.resource_id}"
  policy_type        = "TargetTrackingScaling"
  resource_id        = "${aws_appautoscaling_target.dynamodb_carve_data_write_target.resource_id}"
  scalable_dimension = "${aws_appautoscaling_target.dynamodb_carve_data_write_target.scalable_dimension}"
  service_namespace  = "${aws_appautoscaling_target.dynamodb_carve_data_write_target.service_namespace}"

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "DynamoDBWriteCapacityUtilization"
    }

    target_value = 70
  }
}
