package testhelpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"

	"code.cloudfoundry.org/app-autoscaler/src/autoscaler/helpers/auth"
)

// generateClientCert generates a client certificate with the specified spaceGUID and orgGUID
// included in the organizational unit string.
func GenerateClientCert(orgGUID, spaceGUID string) ([]byte, error) {
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"My Organization"},
			OrganizationalUnit: []string{fmt.Sprintf("space:%s org:%s", spaceGUID, orgGUID)},
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return certPEM, nil
}

func SetXFCCCertHeader(req *http.Request, orgGuid, spaceGuid string) error {
	xfccClientCert, err := GenerateClientCert(orgGuid, spaceGuid)
	if err != nil {
		return err
	}

	cert := auth.NewCert(string(xfccClientCert))

	req.Header.Add("X-Forwarded-Client-Cert", cert.GetXFCCHeader())
	return nil
}
