package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func handler() (string, error) {
	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		log.Println("DOMAIN_NAME is not defined")
		return "", errors.New("DOMAIN_NAME is not defined")
	}
	snsTopicARN := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicARN == "" {
		log.Println("SNS_TOPIC_ARN is not defined")
		return "", errors.New("SNS_TOPIC_ARN is not defined")
	}
	bufferInDays := os.Getenv("BUFFER_IN_DAYS")
	if bufferInDays == "" {
		log.Println("BUFFER_IN_DAYS is not defined")
		return "", errors.New("BUFFER_IN_DAYS is not defined")
	}

	conn, err := tls.Dial(
		"tcp",
		fmt.Sprintf("%s:443", domainName),
		nil)
	if err != nil {
		log.Println(err)
		return "", err
	}

	expirationDate := conn.ConnectionState().PeerCertificates[0].NotAfter

	bufferDays, err := strconv.Atoi(bufferInDays)
	if err != nil {
		log.Println(err)
		return "", err
	}

	now := time.Now()
	buffer := now.Add((time.Hour * 24) * time.Duration(bufferDays))
	certDiffInDays := int(expirationDate.Sub(now).Hours() / 24)

	// Break early when certificate is good
	if buffer.After(expirationDate) {
		return "", nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Println(err)
		return "", err
	}

	client := sns.NewFromConfig(cfg)

	msg := fmt.Sprintf(
		"%s certificate will expire in %d days on %s.",
		domainName,
		certDiffInDays,
		expirationDate.Format(time.RFC822),
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

	output, err := client.Publish(context.Background(), input)
	if err != nil {
		log.Println(err)
		return "", err
	}

	msgID := fmt.Sprintf("MessageID: %s", aws.ToString(output.MessageId))
	return msgID, nil
}

func main() {
	lambda.Start(handler)
}
