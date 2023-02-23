locals {
  function_name = "cert-expiration-checker"
  rule_name     = "cert-checker-trigger"
  runtime       = var.checker_language == "go" ? "go1.x" : var.checker_language == "node" ? "nodejs18.x" : "go1.x"
  go_path       = "${path.module}/certificate_checker_go"
  node_path     = "${path.module}/certificate_checker_node/src"
}

data "archive_file" "lambda_code" {
  type        = "zip"
  source_dir  = local.runtime == "nodejs18.x" ? local.node_path : local.go_path
  output_path = "${path.module}/code.zip"
  excludes    = ["go.mod", "go.sum", "localhost.crt", "localhost.key", "main.go", "main_test.go"]
}

resource "aws_iam_role" "cert_checker_lambda_role" {
  name = "cert-checker-lambda-role"
  tags = var.tags

  assume_role_policy = <<-EOF
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": "sts:AssumeRole",
        "Principal": {
          "Service": "lambda.amazonaws.com"
        }
      }
    ]
  }
  EOF
}

resource "aws_iam_role_policy" "cert_checker_lambda_role_policy" {
  name = "cert-checker-lambda-role-policy"
  role = aws_iam_role.cert_checker_lambda_role.id

  policy = <<-POLICY
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": "sns:Publish",
        "Resource": "${var.sns_topic_arn}"
      }
    ]
  }
  POLICY
}

resource "aws_lambda_function" "cert_expiration_checker" {
  filename      = data.archive_file.lambda_code.output_path
  function_name = local.function_name
  role          = aws_iam_role.cert_checker_lambda_role.arn
  handler       = local.runtime == "nodejs18.x" ? "index.main" : "main"
  runtime       = local.runtime
  tags          = var.tags

  source_code_hash = data.archive_file.lambda_code.output_base64sha256

  environment {
    variables = {
      DOMAIN_NAME    = var.domain_name
      SNS_TOPIC_ARN  = var.sns_topic_arn
      BUFFER_IN_DAYS = var.buffer_in_days
    }
  }
}

resource "aws_cloudwatch_event_rule" "scheduled_trigger" {
  name                = local.rule_name
  schedule_expression = var.schedule_expression
  tags                = var.tags
}

resource "aws_cloudwatch_event_target" "scheduled_target" {
  rule = aws_cloudwatch_event_rule.scheduled_trigger.name
  arn  = aws_lambda_function.cert_expiration_checker.arn
}

resource "aws_lambda_permission" "allow_trigger_invocation" {
  statement_id  = "InvokeLambdaFunction"
  action        = "lambda:InvokeFunction"
  principal     = "events.amazonaws.com"
  function_name = aws_lambda_function.cert_expiration_checker.function_name
  source_arn    = aws_cloudwatch_event_rule.scheduled_trigger.arn
}
