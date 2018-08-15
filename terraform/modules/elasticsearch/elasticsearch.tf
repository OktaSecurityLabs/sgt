provider "aws" {
  profile = "${var.aws_profile}"
  region = "${var.aws_region}"
}

data "aws_caller_identity" "current" {}

resource "aws_elasticsearch_domain_policy" "main" {
  domain_name = "${aws_elasticsearch_domain.sgt-osquery_results.domain_name}"
  access_policies = "${data.aws_iam_policy_document.aws_elasticsearch_domain_policy_doc.json}"
}

data "aws_iam_policy_document" "aws_elasticsearch_domain_policy_doc"{
  statement {
    actions = ["es:ESHttp*"]
    principals {
      type = "AWS"
      identifiers = [
        "${aws_iam_role.authenticated_cognito_role.arn}",
        "${data.aws_caller_identity.current.arn}"
      ]
    }
    resources = ["arn:aws:es:${var.aws_region}:${data.aws_caller_identity.current.account_id}:domain/${var.elasticsearch_domain_name}/*"]
  }
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

resource "aws_iam_role_policy_attachment" "es_cognito_access_role_attach" {
    role       = "${aws_iam_role.es_cognito_access_role.name}"
    policy_arn = "arn:aws:iam::aws:policy/AmazonESCognitoAccess"
}

resource "aws_iam_role" "es_cognito_access_role" {
  name = "${var.es_cognito_access_role_name}",
  assume_role_policy = "${data.aws_iam_policy_document.es_assume_role_policy_doc.json}"
}

data "aws_iam_policy_document" "es_assume_role_policy_doc"{
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type = "Service"
      identifiers = ["es.amazonaws.com"]
    }
  }
}

resource "aws_cognito_identity_pool_roles_attachment" "main" {
  identity_pool_id = "${aws_cognito_identity_pool.main.id}"

  roles {
    "authenticated" = "${aws_iam_role.authenticated_cognito_role.arn}",
    "unauthenticated" = "${aws_iam_role.unauthenticated_cognito_role.arn}"
  }
}

resource "aws_cognito_identity_pool" "main" {
  identity_pool_name               = "${var.identity_pool_name}",
  allow_unauthenticated_identities = false,

}

resource "aws_iam_role_policy" "unauthenticated_cognito_role_policy" {
  name = "unauthenticated_cognito_role_policy"
  role = "${aws_iam_role.unauthenticated_cognito_role.id}"
  policy = "${data.aws_iam_policy_document.unauthenticated_cognito_role_policy_document.json}"
}

data "aws_iam_policy_document" "unauthenticated_cognito_role_policy_document"{
  statement {
    effect = "Allow"
    actions = [
      "mobileanalytics:PutEvents",
      "cognito-sync:*"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "unauthenticated_cognito_role" {
  name = "unauthenticated_cognito_role"
  assume_role_policy = "${data.aws_iam_policy_document.unauthenticated_cognito_assume_role_policy.json}"
}

data "aws_iam_policy_document" "unauthenticated_cognito_assume_role_policy"{
  statement {
    effect = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }
    condition {
      test = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values = ["${aws_cognito_identity_pool.main.id}"]
    }
    condition {
      test = "ForAnyValue:StringLike"
      variable = "cognito-identity.amazonaws.com:amr"
      values = ["unauthenticated"]
    }
  }
}

resource "aws_iam_role_policy" "authenticated_cognito_role_policy" {
  name = "authenticated_cognito_role_policy"
  role = "${aws_iam_role.authenticated_cognito_role.id}"
  policy = "${data.aws_iam_policy_document.authenticated_cognito_role_policy_document.json}"
}

data "aws_iam_policy_document" "authenticated_cognito_role_policy_document"{
  statement {
    effect = "Allow"
    actions = [
      "mobileanalytics:PutEvents",
      "cognito-sync:*",
      "cognito-identity:*"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "authenticated_cognito_role" {
  name = "authenticated_cognito_role"
  assume_role_policy = "${data.aws_iam_policy_document.authenticated_cognito_assume_role_policy.json}"
}

data "aws_iam_policy_document" "authenticated_cognito_assume_role_policy"{
  statement {
    effect = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type = "Federated"
      identifiers = ["cognito-identity.amazonaws.com"]
    }
    condition {
      test = "StringEquals"
      variable = "cognito-identity.amazonaws.com:aud"
      values = ["${aws_cognito_identity_pool.main.id}"]
    }
    condition {
      test = "ForAnyValue:StringLike"
      variable = "cognito-identity.amazonaws.com:amr"
      values = ["authenticated"]
    }
  }
}

resource "aws_cognito_user_pool_domain" "main" {
  domain = "jrichards-kibana-sgt"
  user_pool_id = "${aws_cognito_user_pool.pool.id}"
}

resource "aws_cognito_user_pool" "pool" {
  name = "${var.user_pool_name}",
  mfa_configuration = "OFF"
  admin_create_user_config = {
    unused_account_validity_days = 7,
    allow_admin_create_user_only = "true"
  },
  auto_verified_attributes = [
    "email"
  ],
  password_policy = {
    minimum_length = 12,
    require_lowercase = "true",
    require_numbers = "true",
    require_symbols = "true",
    require_uppercase  = "true"
  }
}
