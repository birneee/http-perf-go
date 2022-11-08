package internal

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

func NewCertPoolWithCert(tlsCertFile string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
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
