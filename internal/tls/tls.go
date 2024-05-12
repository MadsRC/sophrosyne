// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

package tls

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/madsrc/sophrosyne"
)

type KeyType string

const (
	KeyTypeRSA4096 KeyType = "RSA-4096"
	KeyTypeECP224  KeyType = "EC-P224"
	KeyTypeECP256  KeyType = "EC-P256"
	KeyTypeECP384  KeyType = "EC-P384"
	KeyTypeECP521  KeyType = "EC-P521"
	KeyTypeED25519 KeyType = "ED25519"
)

func publicKey(priv any) any {
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

func generateKey(keytype KeyType, randSource io.Reader) (any, error) {
	randSource = ensureRand(randSource)
	var priv any
	var err error
	switch keytype {
	case KeyTypeRSA4096:
		priv, err = rsa.GenerateKey(randSource, 4096)
	case KeyTypeED25519:
		_, priv, err = ed25519.GenerateKey(randSource)
	case KeyTypeECP224:
		priv, err = ecdsa.GenerateKey(elliptic.P224(), randSource)
	case KeyTypeECP256:
		priv, err = ecdsa.GenerateKey(elliptic.P256(), randSource)
	case KeyTypeECP384:
		priv, err = ecdsa.GenerateKey(elliptic.P384(), randSource)
	case KeyTypeECP521:
		priv, err = ecdsa.GenerateKey(elliptic.P521(), randSource)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keytype)
	}

	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ensureRand(randSource io.Reader) io.Reader {
	if randSource == nil {
		randSource = rand.Reader
	}
	return randSource
}

const defaultValidity = 365 * 24 * time.Hour

func generateCert(priv interface{}, hosts []string, validFrom time.Time, validFor time.Duration, isCA bool, randSource io.Reader) ([]byte, error) {
	randSource = ensureRand(randSource)
	var err error
	keyUsage := x509.KeyUsageDigitalSignature
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}
	var notBefore time.Time
	if validFrom.IsZero() {
		notBefore = time.Now()
	} else {
		notBefore = validFrom
	}
	if validFor == 0 {
		validFor = defaultValidity
	}
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(randSource, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Sophrosyne"},
			CommonName:   hosts[0],
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}
	return signCert(&template, &template, publicKey(priv), priv)
}

func signCert(template *x509.Certificate, parent *x509.Certificate, pub any, priv any) ([]byte, error) {
	derBytes, err := x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}
	return derBytes, nil
}

func readPEMFile(path string) (*pem.Block, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	pemfileinfo, _ := file.Stat()
	var size = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode(pembytes)

	return data, nil
}

func readCertificate(path string) ([]byte, error) {
	data, err := readPEMFile(path)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(data.Type, "CERTIFICATE") {
		return nil, fmt.Errorf("PEM data does not contain a certificate. Type is %s", data.Type)
	}

	return data.Bytes, nil
}

// Has to be PKCS8.
func readPrivateKeyPath(path string) (any, error) {
	data, err := readPEMFile(path)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(data.Type, "PRIVATE KEY") {
		return nil, fmt.Errorf("decoded PEM file not as expected. Type is %s", data.Type)
	}

	var me error
	var ret any
	var ecErr error
	var pkcs8err error
	ret, ecErr = x509.ParseECPrivateKey(data.Bytes)
	if ecErr != nil {
		me = errors.Join(me, ecErr)
		ret, pkcs8err = x509.ParsePKCS8PrivateKey(data.Bytes)
		if pkcs8err != nil {
			me = errors.Join(me, pkcs8err)
			return nil, me
		}
	}

	return ret, nil
}

// Create a new [tls.Config] for server use.
//
// The provided config is referenced to determine which settings to set in the returned
// config. If config is nil, a default [tls.Config] is provided.
//
// If the provided randSource is nil, [rand.Reader] will be used.
//
// The following attributes of the provided config are referenced:
//
// Security.TLS.InsecureSkipVerify - if set, the TLS config will be
// configured to no verify certificates.
//
// Security.TLS.KeyPath - Path to an existing TLS key in the filesystem.
// Can be empty, in which case a new private key will be created.
//
// Security.TLS.KeyType - Used to determine what kind of key to generate.
//
// Security.TLS.CertificatePath - Path to an existing X.509 certificate in
// the filesystem. Can be empty, in which case a new certificate will be
// generated.
//
// Server.AdvertisedHost - The value will be used as the common name and first Subject
// Alternative Name of the certificate.
func NewTLSServerConfig(config *sophrosyne.Config, randSource io.Reader) (*tls.Config, error) {
	randSource = ensureRand(randSource)
	if config == nil {
		return newDefaultTLSConfig(), nil
	}
	var priv any
	var err error
	var certBytes []byte
	if config.Security.TLS.KeyPath == "" {
		priv, err = generateKey(KeyType(config.Security.TLS.KeyType), randSource)
	} else {
		priv, err = readPrivateKeyPath(config.Security.TLS.KeyPath)
	}
	if err != nil {
		return nil, err
	}

	if config.Security.TLS.CertificatePath == "" {
		certBytes, err = generateCert(priv, []string{config.Server.AdvertisedHost}, time.Time{}, 0, false, randSource)
	} else {
		certBytes, err = readCertificate(config.Security.TLS.CertificatePath)
	}
	if err != nil {
		return nil, err
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  priv,
	}

	c := newDefaultTLSConfig()
	c.Certificates = []tls.Certificate{cert}
	return c, nil
}

// Create a new [tls.Config] for client use, such as with HTTP calls.
//
// It takes a [sophrosyne.Config] and references it as the source of configuration when
// determining which settings to use in the returned [tls.Config].
//
// If the provided config is nil, a default [tls.Config] is returned.
//
// The following attributes of the provided config are referenced:
//
// Security.TLS.InsecureSkipVerify - if set, the TLS config will be
// configured to not verify certificates.
func NewTLSClientConfig(config *sophrosyne.Config) (*tls.Config, error) {
	c := newDefaultTLSConfig()
	if config == nil {
		return c, nil
	}
	c.InsecureSkipVerify = config.Security.TLS.InsecureSkipVerify

	return c, nil
}

func newDefaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
}
