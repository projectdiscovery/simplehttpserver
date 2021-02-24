// Package sslcert contains a reworked version of https://golang.org/src/crypto/tls/generate_cert.go
package sslcert

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

func pubKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}

func Generate(options Options) (privateKey, publicKey []byte, err error) {
	if options.Host == "" {
		return nil, nil, errors.New("Empty host value")
	}

	var priv interface{}
	switch options.EcdsaCurve {
	case "":
		if options.Ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, options.RSABits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		err = fmt.Errorf("Unrecognized elliptic curve: %q", options.EcdsaCurve)
		return
	}
	if err != nil {
		err = fmt.Errorf("Failed to generate private key: %v", err)
		return
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	var notBefore time.Time
	if len(options.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", options.ValidFrom)
		if err != nil {
			err = fmt.Errorf("Failed to parse creation date: %v", err)
			return
		}
	}

	notAfter := notBefore.Add(options.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		err = fmt.Errorf("Failed to generate serial number: %v", err)
		return
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{options.Organization},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(options.Host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if options.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pubKey(priv), priv)
	if err != nil {
		err = fmt.Errorf("Failed to create certificate: %v", err)
		return
	}

	var pubKeyBuf bytes.Buffer
	pubKeyBufb := bufio.NewWriter(&pubKeyBuf)
	err = pem.Encode(pubKeyBufb, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		err = fmt.Errorf("Failed to write data to cert.pem: %v", err)
		return
	}
	pubKeyBufb.Flush()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		err = fmt.Errorf("Unable to marshal private key: %v", err)
		return
	}
	var privKeyBuf bytes.Buffer
	privKeyBufb := bufio.NewWriter(&privKeyBuf)
	err = pem.Encode(privKeyBufb, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		err = fmt.Errorf("Failed to write data to key.pem: %v", err)
		return
	}
	privKeyBufb.Flush()

	return pubKeyBuf.Bytes(), privKeyBuf.Bytes(), nil
}
