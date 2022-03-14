package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"

	indexer "github.com/dallasmarlow/eks_cert_fingerprint_indexer"
)

func handler() {
	awsSession := session.Must(session.NewSession())
	if err := indexer.Run(
		indexer.NewConfigFromEnv(),
		eks.New(awsSession),
		ssm.New(awsSession),
	); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	lambda.Start(handler)
}
