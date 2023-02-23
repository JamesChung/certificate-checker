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
	ErrDomainName   = errors.New("DOMAIN_NAME is not defined")
	ErrSNSTopicARN  = errors.New("SNS_TOPIC_ARN is not defined")
	ErrBufferInDays = errors.New("BUFFER_IN_DAYS is not defined")
)

type SNSPublish interface {
	Publish(context.Context, *sns.PublishInput, ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type Env struct {
	DomainName   string
	SNSTopicARN  string
	BufferInDays int
}

func getEnv() (Env, error) {
	domainName, err := GetDomainName()
	if err != nil {
		return Env{}, err
	}
	snsTopicARN, err := GetSNSTopicARN()
	if err != nil {
		return Env{}, err
	}
	bufferInDays, err := GetBufferInDays()
	if err != nil {
		return Env{}, err
	}
	return Env{
		DomainName:   domainName,
		SNSTopicARN:  snsTopicARN,
		BufferInDays: bufferInDays,
	}, nil
}

func GetDomainName() (string, error) {
	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		return "", ErrDomainName
	}
	return domainName, nil
}

func GetSNSTopicARN() (string, error) {
	snsTopicARN := os.Getenv("SNS_TOPIC_ARN")
	if snsTopicARN == "" {
		return "", ErrSNSTopicARN
	}
	return snsTopicARN, nil
}

func GetBufferInDays() (int, error) {
	bufferInDays := os.Getenv("BUFFER_IN_DAYS")
	if bufferInDays == "" {
		return 0, ErrBufferInDays
	}
	days, err := parseInt(bufferInDays)
	if err != nil {
		return 0, err
	}
	return days, nil
}

type CertInfo struct {
	Expiration     time.Time
	Buffer         time.Time
	CertDiffInDays int
}

func parseInt(str string) (int, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return i, nil
}

type Dialer interface {
	Dial(network string, addr string, config *tls.Config) (*tls.Conn, error)
}

func getConn(dial Dialer, domainName string) (*tls.Conn, error) {
	conn, err := dial.Dial(
		"tcp",
		fmt.Sprintf("%s:443", domainName),
		&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type Connection interface {
	ConnectionState() tls.ConnectionState
}

func getCertInfo(conn Connection, env Env) CertInfo {
	expirationDate := conn.ConnectionState().PeerCertificates[0].NotAfter
	now := time.Now()
	buffer := now.Add((time.Hour * 24) * time.Duration(env.BufferInDays))
	certDiffInDays := int(expirationDate.Sub(now).Hours() / 24)
	return CertInfo{
		Expiration:     expirationDate,
		Buffer:         buffer,
		CertDiffInDays: certDiffInDays,
	}
}

func constructPubInput(env Env, certInfo CertInfo) *sns.PublishInput {
	msg := fmt.Sprintf(
		"%s certificate will expire in %d days on %s.",
		env.DomainName,
		certInfo.CertDiffInDays,
		certInfo.Expiration.Format(time.RubyDate),
	)
	sub := fmt.Sprintf(
		"%s Certificate Expiring Soon",
		env.DomainName,
	)
	return &sns.PublishInput{
		Message:  aws.String(msg),
		Subject:  aws.String(sub),
		TopicArn: aws.String(env.SNSTopicARN),
	}
}

func pub(ctx context.Context, api SNSPublish, input *sns.PublishInput) (string, error) {
	output, err := api.Publish(ctx, input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("MessageID: %s", aws.ToString(output.MessageId)), nil
}

type TLSDialer struct{}

func (t TLSDialer) Dial(network string, addr string, config *tls.Config) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func handler() (string, error) {
	env, err := getEnv()
	if err != nil {
		return "", err
	}

	dial := TLSDialer{}
	conn, err := getConn(dial, env.DomainName)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	certInfo := getCertInfo(conn, env)

	// Break early when certificate is good
	if certInfo.Expiration.After(certInfo.Buffer) {
		return fmt.Sprintf("%d days left", certInfo.CertDiffInDays), nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return "", err
	}
	client := sns.NewFromConfig(cfg)
	input := constructPubInput(env, certInfo)
	return pub(context.Background(), client, input)
}

func main() {
	lambdaStart(handler)
}
