package main

import (
	"flag"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"

	indexer "github.com/dallasmarlow/eks_cert_fingerprint_indexer"
)

var (
	certificateIndexFlag = flag.Int("cert-reverse-index", indexer.DefaultCertificateIndex, " Reverse index of the certificate to fingerprint within chain, defaults to last cert defined.")
	regionFlag           = flag.String("region", "us-west-2", "AWS region.")
	ssmKeyPrefixFlag     = flag.String("ssm-key-prefix", indexer.DefaultSSMKeyPrefix, "SSM parameter key prefix.")
	ssmOverwriteFlag     = flag.Bool("ssm-overwrite", indexer.DefaultSSMOverwrite, "Overwrite SSM parameters.")
	verifyCertChainFlag  = flag.Bool("verify-cert-chain", indexer.DefaultVerifyCertChain, "Verify TLS certificate chains on read.")
)

func main() {
	flag.Parse()

	var region string
	if envRegion := os.Getenv("AWS_REGION"); envRegion != "" {
		log.Println("setting region from env variable:", envRegion)
		region = envRegion
	} else {
		log.Println("setting region from flag input:", *regionFlag)
		region = *regionFlag
	}

	awsSession := session.Must(
		session.NewSession(
			&aws.Config{
				Region: aws.String(region),
			},
		),
	)

	if err := indexer.Run(
		indexer.Config{
			CertificateIndex: *certificateIndexFlag,
			SSMKeyPrefix:     *ssmKeyPrefixFlag,
			SSMOverwrite:     *ssmOverwriteFlag,
			VerifyCertChain:  *verifyCertChainFlag,
		},
		eks.New(awsSession),
		ssm.New(awsSession),
	); err != nil {
		log.Fatalln(err)
	}
}
