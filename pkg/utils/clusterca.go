package utils

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	openshiftclientset "github.com/openshift/client-go/config/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	apiserverConfigName      = "cluster"
	openshiftConfigNamespace = "openshift-config"
)

// getKubeAPIServerSecretName iterate through all namespacedCertificates
// returns the first one which has a name matches the given dnsName
func getKubeAPIServerSecretName(ocpClient openshiftclientset.Interface, dnsName string) (string, error) {
	apiserver, err := ocpClient.ConfigV1().APIServers().Get(context.Background(), apiserverConfigName, v1.GetOptions{})
	if err != nil {
		return "", err
	}

	// iterate through all namedcertificates
	for _, namedCert := range apiserver.Spec.ServingCerts.NamedCertificates {
		for _, name := range namedCert.Names {
			if strings.EqualFold(name, dnsName) {
				return namedCert.ServingCertificate.Name, nil
			}
		}
	}

	return "", fmt.Errorf("Failed to get ServingCerts match name: %s", dnsName)
}

// getKubeAPIServerCertificate looks for secret in openshift-config namespace, and returns tls.crt
func getKubeAPIServerCertificate(kubeClient kubernetes.Interface, secretName string) ([]byte, error) {
	secret, err := kubeClient.CoreV1().Secrets(openshiftConfigNamespace).Get(context.Background(), secretName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if secret.Type != corev1.SecretTypeTLS {
		return nil, fmt.Errorf(
			"secret %s/%s should have type=kubernetes.io/tls",
			openshiftConfigNamespace,
			secretName,
		)
	}
	res, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find data[tls.crt] in secret %s/%s",
			openshiftConfigNamespace,
			secretName,
		)
	}
	return res, nil
}

func GetClusterCa(ocpClient openshiftclientset.Interface, kubeClient kubernetes.Interface, kubeAPIServer string) ([]byte, error) {
	u, err := url.Parse(kubeAPIServer)
	if err != nil {
		return []byte{}, err
	}
	apiServerCertSecretName, err := getKubeAPIServerSecretName(ocpClient, u.Hostname())
	if err != nil {
		return nil, err
	}

	apiServerCert, err := getKubeAPIServerCertificate(kubeClient, apiServerCertSecretName)
	if err != nil {
		return nil, err
	}
	return []byte(apiServerCert), nil
}
