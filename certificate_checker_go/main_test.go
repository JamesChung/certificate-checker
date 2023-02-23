package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	domain = "example.com"
	arn    = "arn:aws:sns:us-east-1:000000000000:blah"
	bid    = "30"
)

func mockLambdaStart(i interface{}) {}

func resetEnv() {
	os.Setenv("DOMAIN_NAME", "")
	os.Setenv("SNS_TOPIC_ARN", "")
	os.Setenv("BUFFER_IN_DAYS", "")
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

func TestEnv(t *testing.T) {
	// Testing DomainName
	t.Run("Test DomainName success", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", bid)
		env, err := getEnv()
		if err != nil {
			t.Fail()
		}
		if env.DomainName != domain {
			t.Fail()
		}
	})
	t.Run("Test DomainName failure", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", "")
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", bid)
		env, err := getEnv()
		if !errors.Is(err, ErrDomainName) {
			t.Fail()
		}
		if env.DomainName != "" {
			t.Fail()
		}
	})

	// Testing SNSTopicARN
	t.Run("Test SNSTopicARN success", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", bid)
		env, err := getEnv()
		if err != nil {
			t.Fail()
		}
		if env.SNSTopicARN != arn {
			t.Fail()
		}
	})
	t.Run("Test SNSTopicARN failure", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", "")
		os.Setenv("BUFFER_IN_DAYS", bid)
		env, err := getEnv()
		if !errors.Is(err, ErrSNSTopicARN) {
			t.Fail()
		}
		if env.SNSTopicARN != "" {
			t.Fail()
		}
	})

	// Testing BufferInDays
	t.Run("Test BufferInDays success", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", bid)
		env, err := getEnv()
		if err != nil {
			t.Fail()
		}
		if env.BufferInDays != 30 {
			t.Fail()
		}
	})
	t.Run("Test BufferInDays empty failure", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", "")
		env, err := getEnv()
		if !errors.Is(err, ErrBufferInDays) {
			t.Fail()
		}
		if env.SNSTopicARN != "" {
			t.Fail()
		}
	})
	t.Run("Test BufferInDays parse int failure", func(t *testing.T) {
		resetEnv()
		os.Setenv("DOMAIN_NAME", domain)
		os.Setenv("SNS_TOPIC_ARN", arn)
		os.Setenv("BUFFER_IN_DAYS", "a")
		env, err := getEnv()
		if err == nil {
			t.Fail()
		}
		if env.SNSTopicARN != "" {
			t.Fail()
		}
	})
}

type MockDialSuccess struct{}

func (d MockDialSuccess) Dial(network string, addr string, config *tls.Config) (*tls.Conn, error) {
	return &tls.Conn{}, nil
}

type MockDialFail struct{}

func (d MockDialFail) Dial(network string, addr string, config *tls.Config) (*tls.Conn, error) {
	return nil, errors.New("fail")
}

func Test_getConn(t *testing.T) {
	t.Run("get connection success", func(t *testing.T) {
		dial := MockDialSuccess{}
		_, err := getConn(dial, "example.com")
		if err != nil {
			t.Fail()
		}
	})
	t.Run("get connection fail", func(t *testing.T) {
		dial := MockDialFail{}
		_, err := getConn(dial, "example.com")
		if err == nil {
			t.Fail()
		}
	})
}

type MockConn struct{}

func (c MockConn) ConnectionState() tls.ConnectionState {
	return tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{
				NotAfter: time.Now(),
			},
		},
	}
}

func Test_getCertInfo(t *testing.T) {
	conn := MockConn{}
	resetEnv()
	os.Setenv("DOMAIN_NAME", domain)
	os.Setenv("SNS_TOPIC_ARN", arn)
	os.Setenv("BUFFER_IN_DAYS", bid)
	env, _ := getEnv()
	certInfo := getCertInfo(conn, env)
	if certInfo.Buffer == time.Now() {
		t.Fail()
	}
}

func Test_constructPubInput(t *testing.T) {
	conn := MockConn{}
	resetEnv()
	os.Setenv("DOMAIN_NAME", domain)
	os.Setenv("SNS_TOPIC_ARN", arn)
	os.Setenv("BUFFER_IN_DAYS", bid)
	env, _ := getEnv()
	certInfo := getCertInfo(conn, env)
	input := constructPubInput(env, certInfo)
	if input == nil {
		t.Fail()
	}
}

type mockSNSSuccess struct{}

func (m mockSNSSuccess) Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return &sns.PublishOutput{}, nil
}

type mockSNSFail struct{}

func (m mockSNSFail) Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return nil, errors.New("failed")
}

func Test_pub(t *testing.T) {
	t.Run("Publish successfully", func(t *testing.T) {
		resetEnv()
		m := mockSNSSuccess{}
		_, err := pub(context.Background(), m, &sns.PublishInput{})
		if err != nil {
			t.Fail()
		}
	})
	t.Run("Fail to publish", func(t *testing.T) {
		resetEnv()
		m := mockSNSFail{}
		_, err := pub(context.Background(), m, &sns.PublishInput{})
		if err == nil {
			t.Fail()
		}
	})
}

func Test_handler(t *testing.T) {
	t.Run("env failure", func(t *testing.T) {
		resetEnv()
		_, err := handler()
		if err == nil {
			t.Fail()
		}
	})
	t.Run("dial", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "hello")
			})
			fmt.Println(http.ListenAndServeTLS(":443", "localhost.crt", "localhost.key", nil))
		}()
		go func() {
			defer wg.Done()
			// Success
			resetEnv()
			os.Setenv("DOMAIN_NAME", "localhost")
			os.Setenv("SNS_TOPIC_ARN", arn)
			os.Setenv("BUFFER_IN_DAYS", bid)
			_, err := handler()
			if err != nil {
				t.Fail()
			}
			// Fail
			resetEnv()
			os.Setenv("DOMAIN_NAME", "localhost")
			os.Setenv("SNS_TOPIC_ARN", arn)
			os.Setenv("BUFFER_IN_DAYS", "1000")
			_, err = handler()
			if err == nil {
				t.Fail()
			}
		}()
		wg.Wait()
	})
}
