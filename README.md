# certificate-checker
Terraform module to create a scheduled event to check a given endpoint certificate expiration date and notify via SNS

## Usage

> This will publish a message to the SNS topic when the certificate on the endpoint is about to expire in 60 days or less.

```terraform
module "certificate_checker" {
    source = "github.com/JamesChung/certificate-checker"

    sns_topic_arn  = "arn:aws:sns:us-east-1:000000000000:my-topic"
    domain_name    = "google.com"
    buffer_in_days = 60
}
```

## Inputs

|Name|Description|Type|Default|Required|
|:-:|:-:|:-:|:-:|:-:|
|sns_topic_arn|ARN of the SNS topic the lambda will broadcast to|`string`||Yes|
|domain_name|The domain name of the endpoint to check certificate expiration|`string`||Yes|
|buffer_in_days|Buffer in days when to start alerts|`number`||Yes|
|schedule_expression|The scheduled rate at which to check the certificate ([Expression Reference](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html))|`string`|`"rate(1 day)"`|No|
|tags|A map of tags to add to all resources|`map(string)`|`{}`|No|

## Outputs

|Name|Description|Type|
|:-:|:-:|:-:|
|lambda_arn|ARN of scheduled lambda used to process certificate expiration|`string`|
|lambda_name|Name of scheduled lambda used to process certificate expiration|`string`|
|event_name|Name of the scheduled event|`string`|
|event_arn|ARN of the scheduled event|`string`|
