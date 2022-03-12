package eks_cert_fingerprint_indexer

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	errCnt            = 0
	errInvalidScheme  = errors.New("Invalid URL scheme, supported schemes: [https].")
	errPartialFailure = errors.New("Indexer experianced failures during run.")
)

func clusters(svc *eks.EKS) ([]string, error) {
	var clusters []string

	log.Println("listing clusters")
	err := svc.ListClustersPages(
		&eks.ListClustersInput{},
		func(resp *eks.ListClustersOutput, lastPage bool) bool {
			for _, cluster := range resp.Clusters {
				clusters = append(clusters, *cluster)
			}
			return !lastPage
		},
	)

	return clusters, err
}

func clusterOidcIssuerURL(cluster string, svc *eks.EKS) (string, error) {
	log.Println("describing cluster:", cluster)
	response, err := svc.DescribeCluster(
		&eks.DescribeClusterInput{
			Name: aws.String(cluster),
		},
	)
	if err != nil {
		return "", err
	}

	return *response.Cluster.Identity.Oidc.Issuer, nil
}

func fingerprintTlsCert(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha1.Sum(cert.Raw))
}

func readTlsCerts(endpointURL string, verifyCertChain bool) ([]*x509.Certificate, error) {
	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "https" {
		return nil, errInvalidScheme
	}
	if parsedURL.Port() == "" {
		parsedURL.Host += ":443"
	}

	log.Println("reading certificates from endpoint:", endpointURL)
	conn, err := tls.Dial(
		"tcp",
		parsedURL.Host,
		&tls.Config{
			InsecureSkipVerify: !verifyCertChain,
		},
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, nil
}

func Run(config Config, eksService *eks.EKS, ssmService *ssm.SSM) error {
	eksClusters, err := clusters(eksService)
	if err != nil {
		log.Println("error - unable to list EKS clusters, err:", err)
		return err
	}

	for _, cluster := range eksClusters {
		ssm_key := path.Join(config.SSMKeyPrefix, cluster)
		if !config.SSMOverwrite {
			log.Println("checking for existing SSM parameter at path:", ssm_key)
			_, err := ssmService.GetParameter(&ssm.GetParameterInput{
				Name:           aws.String(ssm_key),
				WithDecryption: aws.Bool(false),
			})
			if err == nil {
				log.Println("SSM parameter already exists:", ssm_key)
				continue
			} else {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() != "ParameterNotFound" {
						log.Println("error - unable to get SSM parameter:", ssm_key, "err:", err)
						errCnt += 1
						continue
					}
				}
			}
		}

		isserUrl, err := clusterOidcIssuerURL(cluster, eksService)
		if err != nil {
			log.Println("error - unable to detect cluster OIDC issuer URL for cluster:", cluster)
			errCnt += 1
			continue
		}

		certs, err := readTlsCerts(isserUrl, config.VerifyCertChain)
		if err != nil {
			log.Println("error - unable to read certificates from endpoint:", isserUrl, "err:", err)
			errCnt += 1
			continue
		}

		certIndex := len(certs) - config.CertificateIndex - 1
		certFingerprint := fingerprintTlsCert(certs[certIndex])

		log.Println("setting SSM parameter:", ssm_key)
		if _, err := ssmService.PutParameter(&ssm.PutParameterInput{
			Name:      aws.String(ssm_key),
			Overwrite: aws.Bool(config.SSMOverwrite),
			Type:      aws.String("String"),
			Value:     aws.String(certFingerprint),
		}); err != nil {
			log.Println("error - unable to set SSM parameter:", ssm_key, "err:", err)
			errCnt += 1
			continue
		}
	}

	if errCnt > 0 {
		return errPartialFailure
	}

	return nil
}
