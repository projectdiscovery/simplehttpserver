package sslcert

import "time"

type Options struct {
	Host         string // Comma-separated hostnames and IPs to generate a certificate for")
	Organization string
	ValidFrom    string        // Creation date formatted as Jan 1 15:04:05 2011
	ValidFor     time.Duration // 365*24*time.Hour Duration that certificate is valid for
	IsCA         bool          // whether this cert should be its own Certificate Authority
	RSABits      int           // 2048 Size of RSA key to generate. Ignored if --ecdsa-curve is set
	EcdsaCurve   string        // ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521
	Ed25519Key   bool          // Generate an Ed25519 key
}

var DefaultOptions = Options{
	ValidFor:     time.Duration(365 * 24 * time.Hour),
	IsCA:         false,
	RSABits:      2048,
	Organization: "Acme Co",
}
