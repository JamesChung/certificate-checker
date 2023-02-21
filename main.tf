locals {
  function_name = "cert-expiration-checker"
  rule_name     = "cert-checker-trigger"
  runtime       = "go1.x"
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
  filename      = "certificate_checker_go/main.zip"
  function_name = local.function_name
  role          = aws_iam_role.cert_checker_lambda_role.arn
  handler       = "main"
  runtime       = local.runtime
  tags          = var.tags

  source_code_hash = filesha256("certificate_checker_go/main.zip")

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
