package eks_cert_fingerprint_indexer

import (
	"log"
	"os"
	"regexp"
	"strconv"
)

const (
	CertificateIndexEnvVar = "CERT_INDEX"
	SSMKeyPrefixEnvVar     = "SSM_KEY_PFX"
	SSMOverwriteEnvVar     = "SSM_OVERWRITE"
	VerifyCertChainEnvVar  = "VERIFY_CERTS"

	DefaultCertificateIndex = 0
	DefaultSSMKeyPrefix     = "/eks_cluster_oidc_fingerprints/"
	DefaultSSMOverwrite     = false
	DefaultVerifyCertChain  = true

	ssmPathPattern = "^/[0-9A-Za-z_/.-]+/$"
)

var (
	ssmRegex = regexp.MustCompile(ssmPathPattern)
)

type Config struct {
	CertificateIndex int
	SSMKeyPrefix     string
	SSMOverwrite     bool
	VerifyCertChain  bool
}

func NewConfig() Config {
	return Config{
		CertificateIndex: DefaultCertificateIndex,
		SSMKeyPrefix:     DefaultSSMKeyPrefix,
		SSMOverwrite:     DefaultSSMOverwrite,
		VerifyCertChain:  DefaultVerifyCertChain,
	}
}

func NewConfigFromEnv() Config {
	config := NewConfig()

	if certIdx := os.Getenv(CertificateIndexEnvVar); certIdx != "" {
		if certIdxVal, err := strconv.Atoi(certIdx); err == nil {
			config.CertificateIndex = certIdxVal
		} else {
			log.Println(
				"error - unable to parse env var:",
				CertificateIndexEnvVar,
				"value:",
				certIdx,
				"err:",
				err,
			)
		}
	}

	if ssmPfx := os.Getenv(SSMKeyPrefixEnvVar); ssmPfx != "" {
		if ssmRegex.MatchString(ssmPfx) {
			config.SSMKeyPrefix = ssmPfx
		} else {
			log.Println(
				"error - invalid env var value:",
				SSMKeyPrefixEnvVar,
				"value:",
				ssmPfx,
				"validation pattern:",
				ssmPathPattern,
			)
		}
	}

	if ssmOverwrite := os.Getenv(SSMOverwriteEnvVar); ssmOverwrite != "" {
		if ssmOverwriteVal, err := strconv.ParseBool(ssmOverwrite); err == nil {
			config.SSMOverwrite = ssmOverwriteVal
		} else {
			log.Println(
				"error - unable to parse env var:",
				SSMOverwriteEnvVar,
				"value:",
				ssmOverwrite,
				"err:",
				err,
			)
		}
	}

	if verifyCerts := os.Getenv(VerifyCertChainEnvVar); verifyCerts != "" {
		if verifyCertsVal, err := strconv.ParseBool(verifyCerts); err == nil {
			config.VerifyCertChain = verifyCertsVal
		} else {
			log.Println(
				"error - unable to parse env var:",
				VerifyCertChainEnvVar,
				"value:",
				verifyCerts,
				"err:",
				err,
			)
		}
	}

	return config
}
