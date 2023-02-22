package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func mockLambdaStart(i interface{}) {}

func resetEnv() {
	os.Setenv("DOMAIN_NAME", "")
	os.Setenv("SNS_TOPIC_ARN", "")
	os.Setenv("BUFFER_IN_DAYS", "")
}

func Test_handler(t *testing.T) {
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

func Test_main(t *testing.T) {
	t.Run("main", func(t *testing.T) {
		var testLambdaStart = lambdaStart
		defer func() {
			lambdaStart = testLambdaStart
		}()
		lambdaStart = mockLambdaStart
		main()
	})
}

func Test_getDomainName(t *testing.T) {
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

func Test_getSNSTopicARN(t *testing.T) {
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

func Test_getBufferInDays(t *testing.T) {
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

type mockSNS struct{}

func (m mockSNS) Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{}, nil
}

func Test_pub(t *testing.T) {
	resetEnv()
	m := mockSNS{}
	_, err := pub(context.Background(), m, &sns.PublishInput{})
	if err != nil {
		t.Fail()
	}
}
