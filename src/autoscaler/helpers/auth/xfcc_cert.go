package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type Cert struct {
	FullChainPem string
	Sha256       [32]byte
	Base64       string
}

func NewCert(fullChainPem string) *Cert {
	block, _ := pem.Decode([]byte(fullChainPem))
	if block == nil {
		return nil
	}
	return &Cert{
		FullChainPem: fullChainPem,
		Sha256:       sha256.Sum256(block.Bytes),
		Base64:       base64.StdEncoding.EncodeToString(block.Bytes),
	}
}

func (c *Cert) GetXFCCHeader() string {
	return fmt.Sprintf("Hash=%x;Cert=%s", c.Sha256, c.Base64)
}
