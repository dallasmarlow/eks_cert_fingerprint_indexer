# eks_cert_fingerprint_indexer

This repository contains a utility program which distributes TLS certificate SHA1 fingerprints for [AWS](https://aws.amazon.com/) [EKS](https://aws.amazon.com/eks/) cluster identity [OIDC](https://openid.net/connect/) issuer endpoints via [AWS SSM parameters](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html).

## Why is this a thing?

AWS EKS clusters [use](https://aws.amazon.com/blogs/containers/introducing-oidc-identity-provider-authentication-amazon-eks/) OpenID Connect Providers when granting [AWS IAM](https://aws.amazon.com/iam/) credentials to K8S entities (namespaces, service accounts). When [creating](https://docs.aws.amazon.com/IAM/latest/APIReference/API_CreateOpenIDConnectProvider.html) a new OIDC provider to use with an existing EKS cluster you must supply a certificate fingerprint of the OIDC issuer that was used when the EKS cluster was created. 

AWS provides the following documentation with instructions on how to obtain the certificate fingerprint manually: <https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html>.

Typically when managing EKS clusters using [Terraform](https://www.terraform.io/) you can use the [TLS provider](https://registry.terraform.io/providers/hashicorp/tls/latest/docs) to fetch the certificate fingerprint associated with the OIDC issuer associated with a given EKS cluster directly. Example:

```
data "tls_certificate" "cluster_issuer" {
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks_cluster" {
  client_id_list = [
    "sts.amazonaws.com",
  ]
  thumbprint_list = [
    data.tls_certificate.cluster_issuer.certificates[0].sha1_fingerprint,
  ]
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}
```

The Terraform code above provides the same functionality as the utility program within this repo in a more direct manner, but this approach will only work if the Terraform process has direct access to AWS IP addresses used with the OIDC hosted services. 

In the event that the Terraform process uses a forward proxy (e.g.: Squid) or other systems which limit internet egress access to specific DNS zones, the approach above will fail due to the Terraform TLS provider [tls_certificate data source](https://registry.terraform.io/providers/hashicorp/tls/latest/docs/data-sources/tls_certificate) implementation which will attempt to open a TCP connection to the host IP address directly.

The `eks_cert_fingerprint_indexer` program is designed to provide an alternative solution for people who find themselves in the situation described above.

## What is this thing?

Simply put, the `eks_cert_fingerprint_indexer` program reads EKS OIDC issuer certificates, generates SHA1 fingerprints and stores them in AWS SSM parameters. This allows other provisioning systems (e.g.: [Terraform](https://www.terraform.io/), [CloudFormation](https://aws.amazon.com/cloudformation/)) to fetch the fingerprint values and use them when creating new AWS OIDC providers. Terraform example:

```
data "aws_ssm_parameter" "eks_cluster_oidc_fingerprint" {
  name = "/eks_cluster_oidc_fingerprints/example"
}

resource "aws_iam_openid_connect_provider" "eks_cluster" {
  client_id_list = [
    "sts.amazonaws.com",
  ]
  thumbprint_list = [
    data.aws_ssm_parameter.eks_cluster_oidc_fingerprint.value,
  ]
  url = aws_eks_cluster.example.identity[0].oidc[0].issuer
}
```

The `eks_cert_fingerprint_indexer` program performs the following actions when executed:
  - List EKS clusters within a configured AWS region.
  - Describe each EKS cluster to fetch the OIDC issuer URL associated.
  - Establish TLS connection to each OIDC issuer URL to read certificates.
  - Generate SHA1 "fingerprint" of the configured certificate index (default is last certificate defined) within the certificate chain.
  - Create a SSM parameter using the configured key prefix suffixed by `/<EKS cluster name>`.
