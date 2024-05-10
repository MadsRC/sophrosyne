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
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/madsrc/sophrosyne"
)

var (
	host      = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
	validFrom = flag.String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	validFor  = flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	isCA      = flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
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

func publicKey(priv interface{}) interface{} {
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

func generateKey(keytype KeyType, randSource io.Reader) (interface{}, error) {
	var priv interface{}
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

func generateCert(priv interface{}, randSource io.Reader) ([]byte, error) {
	var err error
	keyUsage := x509.KeyUsageDigitalSignature
	if _, isRSA := priv.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}
	var notBefore time.Time
	if len(*validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", *validFrom)
		if err != nil {
			log.Fatalf("Failed to parse creation date: %v", err)
		}
	}
	notAfter := notBefore.Add(*validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(randSource, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hosts := strings.Split(*host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	if *isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}
	derBytes, err := x509.CreateCertificate(randSource, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
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
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(pembytes)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode([]byte(pembytes))

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
func readPrivateKeyPath(path string) (interface{}, error) {
	data, err := readPEMFile(path)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(data.Type, "PRIVATE KEY") {
		return nil, fmt.Errorf("decoded PEM file not as expected. Type is %s", data.Type)
	}

	return x509.ParsePKCS8PrivateKey(data.Bytes)
}

func NewTLSServerConfig(config *sophrosyne.Config, randSource io.Reader) (*tls.Config, error) {
	var priv interface{}
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
		certBytes, err = generateCert(priv, randSource)
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

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func NewTLSClientConfig(config *sophrosyne.Config) (*tls.Config, error) {
	c := &tls.Config{}
	if config.Security.TLS.InsecureSkipVerify {
		c.InsecureSkipVerify = true
	}

	return c, nil
}
