variable "sns_topic_arn" {
  description = "(Required) ARN of the SNS topic the lambda will broadcast to"
  type        = string
}

variable "domain_name" {
  description = "(Required) The domain name of the endpoint to check certificate expiration"
  type        = string
}

variable "buffer_in_days" {
  description = "(Required) Buffer in days on when to start alerts"
  type        = number
}

variable "schedule_expression" {
  description = "(Optional) The scheduled rate at which to check the certificate"
  type        = string
  default     = "rate(1 day)"
}

variable "tags" {
  description = "(Optional) A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
