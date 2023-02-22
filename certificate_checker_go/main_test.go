package main

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func mockLambdaStart(i interface{}) {}

func resetEnv() {
	os.Setenv("DOMAIN_NAME", "")
	os.Setenv("SNS_TOPIC_ARN", "")
	os.Setenv("BUFFER_IN_DAYS", "")
}

func TestHandler(t *testing.T) {
	t.Run("Domain Name", func(t *testing.T) {
		v, err := handler()
		if err != DomainNameErr && v != "" {
			t.Fail()
		}
	})
	t.Run("SNS Topic ARN", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "example.com")
		v, err := handler()
		if err != SNSTopicARNErr && v != "" {
			t.Fail()
		}
	})
	t.Run("Buffer in Days", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "example.com")
		os.Setenv("SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:blah")
		v, err := handler()
		if err != BufferInDaysErr && v != "" {
			t.Fail()
		}
	})
	t.Run("Fail on incorrect BUFFER_IN_DAYS value", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "example.com")
		os.Setenv("SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:blah")
		os.Setenv("BUFFER_IN_DAYS", "fail")
		_, err := handler()
		if !errors.Is(err, strconv.ErrSyntax) {
			t.Fail()
		}
	})

	t.Run("Run", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "example.com")
		os.Setenv("SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:blah")
		os.Setenv("BUFFER_IN_DAYS", "30")
		_, err := handler()
		if err == nil {
			t.Fail()
		}
	})
}

func TestMain(t *testing.T) {
	t.Run("main", func(t *testing.T) {
		var testLambdaStart = lambdaStart
		defer func() {
			lambdaStart = testLambdaStart
		}()
		lambdaStart = mockLambdaStart
		main()
	})
}

func TestGetDomainName(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "")
		_, err := getDomainName()
		if err == nil {
			t.Fail()
		}
	})

	t.Run("good", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "google.com")
		_, err := getDomainName()
		if err != nil {
			t.Fail()
		}
	})
}

func TestGetSNSTopicARN(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		resetEnv()
		os.Setenv("SNS_TOPIC_ARN", "")
		_, err := getSNSTopicARN()
		if err == nil {
			t.Fail()
		}
	})

	t.Run("good", func(t *testing.T) {
		resetEnv()
		os.Setenv("SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:blah")
		_, err := getSNSTopicARN()
		if err != nil {
			t.Fail()
		}
	})
}

func TestGetBufferInDays(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		resetEnv()
		os.Setenv("BUFFER_IN_DAYS", "")
		_, err := getBufferInDays()
		if err == nil {
			t.Fail()
		}
	})

	t.Run("good", func(t *testing.T) {
		resetEnv()
		os.Setenv("BUFFER_IN_DAYS", "30")
		_, err := getBufferInDays()
		if err != nil {
			t.Fail()
		}
	})
}

func TestPublishMessage(t *testing.T) {
	t.Run("Publish Fail", func(t *testing.T) {
		_, err := publishMessage(&sns.PublishInput{
			Message: aws.String(""),
			Subject: aws.String(""),
		})
		if err == nil {
			t.Fail()
		}
	})
}
