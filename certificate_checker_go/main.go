package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	lambdaStart     = lambda.Start
	DomainNameErr   = errors.New("DOMAIN_NAME is not defined")
	SNSTopicARNErr  = errors.New("SNS_TOPIC_ARN is not defined")
	BufferInDaysErr = errors.New("BUFFER_IN_DAYS is not defined")
)

type SNSPublish interface {
	Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func getDomainName() (string, error) {
	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		return "", DomainNameErr
	}
	return domainName, nil
}

func getSNSTopicARN() (string, error) {
	snsTopicARN := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicARN == "" {
		return "", SNSTopicARNErr
	}
	return snsTopicARN, nil
}

func getBufferInDays() (string, error) {
	bufferInDays := os.Getenv("BUFFER_IN_DAYS")
	if bufferInDays == "" {
		return "", BufferInDaysErr
	}
	return bufferInDays, nil
}

func pub(ctx context.Context, api SNSPublish, input *sns.PublishInput) (string, error) {
	output, err := api.Publish(ctx, input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MessageID: %s", aws.ToString(output.MessageId)), nil
}

func handler() (string, error) {
	domainName, err := getDomainName()
	if err != nil {
		return "", err
	}
	snsTopicARN, err := getSNSTopicARN()
	if err != nil {
		return "", err
	}
	bufferInDays, err := getBufferInDays()
	if err != nil {
		return "", err
	}

	conn, _ := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:443", domainName),
		nil)

	expirationDate := conn.ConnectionState().PeerCertificates[0].NotAfter

	bufferDays, err := strconv.Atoi(bufferInDays)
	if err != nil {
		return "", err
	}

	now := time.Now()
	buffer := now.Add((time.Hour * 24) * time.Duration(bufferDays))
	certDiffInDays := int(expirationDate.Sub(now).Hours() / 24)

	// Break early when certificate is good
	if buffer.After(expirationDate) {
		return "", nil
	}

	msg := fmt.Sprintf(
		"%s certificate will expire in %d days on %s.",
		domainName,
		certDiffInDays,
		expirationDate.Format(time.RubyDate),
	)
	sub := fmt.Sprintf(
		"%s Certificate Expiring Soon",
		domainName,
	)
	input := &sns.PublishInput{
		Message:  aws.String(msg),
		Subject:  aws.String(sub),
		TopicArn: aws.String(snsTopicARN),
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return "", err
	}
	client := sns.NewFromConfig(cfg)
	msgID, err := pub(context.Background(), client, input)
	if err != nil {
		return "", err
	}

	return msgID, nil
}

func main() {
	lambdaStart(handler)
}
