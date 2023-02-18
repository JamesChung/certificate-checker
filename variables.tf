variable "sns_topic_arn" {
  description = "ARN of the SNS topic the lambda will broadcast to"
  type        = string
}

variable "domain_name" {
  description = "The domain name of the endpoint to check certificate expiration"
  type        = string
}

variable "buffer_in_days" {
  description = "Buffer in days on when to start alerts"
  type        = number
}

variable "schedule_expression" {
  description = "The scheduled rate at which to check the certificate"
  type        = string
  default     = "rate(1 day)"
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}
