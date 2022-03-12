package eks_cert_fingerprint_indexer

const (
	DefaultCertificateIndex = 0
	DefaultSSMKeyPrefix     = "/eks_cluster_oidc_fingerprints/"
	DefaultSSMOverwrite     = false
	DefaultVerifyCertChain  = true
)

type Config struct {
	CertificateIndex int
	SSMKeyPrefix     string
	SSMOverwrite     bool
	VerifyCertChain  bool
}
