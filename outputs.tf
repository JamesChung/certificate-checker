output "lambda_arn" {
  value = aws_lambda_function.cert_expiration_checker.arn
}

output "lambda_name" {
  value = aws_lambda_function.cert_expiration_checker.function_name
}

output "event_name" {
  value = aws_cloudwatch_event_rule.scheduled_trigger.name
}

output "event_arn" {
  value = aws_cloudwatch_event_rule.scheduled_trigger.arn
}
