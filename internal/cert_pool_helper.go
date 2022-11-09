package internal

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

func SystemCertPoolWithAdditionalCert(tlsCertFile string) (*x509.CertPool, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system cert pool: %v", err)
	}

	caCertRaw, err := ioutil.ReadFile(tlsCertFile)
	if err != nil {
		return nil, err
	}

	ok := certPool.AppendCertsFromPEM(caCertRaw)
	if !ok {
		return nil, fmt.Errorf("failed to add certificate to pool")
	}
	return certPool, nil
}
