package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
)

func TestNewTLSClientConfig(t *testing.T) {
	type args struct {
		config *sophrosyne.Config
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
		{
			name: "empty config",
			args: args{
				&sophrosyne.Config{},
			},
			want: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
		{
			name: "config with insecure tls",
			args: args{
				&sophrosyne.Config{
					Security: sophrosyne.SecurityConfig{
						TLS: sophrosyne.TLSConfig{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			want: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
		{
			name: "nil config",
			args: args{},
			want: newDefaultTLSConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTLSClientConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNewTLSServerConfig(t *testing.T) {
	type args struct {
		config     *sophrosyne.Config
		randSource io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successfull call",
			args: args{
				config: &sophrosyne.Config{
					Security: sophrosyne.SecurityConfig{
						TLS: sophrosyne.TLSConfig{
							KeyType: "EC-P384",
						},
					},
				},
			},
		},
		{
			name: "empty config",
			args: args{
				config: &sophrosyne.Config{},
			},
			wantErr: true,
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
		},
		{
			name: "bad TLS key path",
			args: args{
				config: &sophrosyne.Config{
					Security: sophrosyne.SecurityConfig{
						TLS: sophrosyne.TLSConfig{
							KeyPath: "/doesnotexist",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "bad certificate path",
			args: args{
				config: &sophrosyne.Config{
					Security: sophrosyne.SecurityConfig{
						TLS: sophrosyne.TLSConfig{
							KeyType:         "EC-P384",
							CertificatePath: "/doesnotexist",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTLSServerConfig(tt.args.config, tt.args.randSource)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			if tt.args.config != nil {
				require.NotNil(t, got.Certificates)
				require.Len(t, got.Certificates, 1)
			}

		})
	}
}

func checkCert(t *testing.T, cert *x509.Certificate, args generateCertArgs) {
	t.Helper()
	require.NotNil(t, cert)

	_, ok := args.priv.(*rsa.PrivateKey)
	if ok && !args.isCA {
		require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment)
	} else if ok && args.isCA {
		require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment|x509.KeyUsageCertSign)
		require.True(t, cert.IsCA)
	} else if !ok && args.isCA {
		require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign)
		require.True(t, cert.IsCA)
	} else {
		require.Equal(t, cert.KeyUsage, x509.KeyUsageDigitalSignature)
	}

	requirePrivSignedCert(t, args.priv, cert)

	// Validate that if validFrom is zero, the certificate should have used time.Now.
	if args.validFrom.IsZero() {
		require.Equal(t, cert.NotBefore, time.Now().UTC().Truncate(time.Second))
	} else {
		require.Equal(t, args.validFrom.UTC().Truncate(time.Second), cert.NotBefore)
	}

	vFor := defaultValidity
	if args.validFor != 0 {
		vFor = args.validFor
	}
	require.Equal(t, vFor, cert.NotAfter.Sub(cert.NotBefore))

	require.Equal(t, args.hosts[0], cert.Subject.CommonName)
	require.Equal(t, "Sophrosyne", cert.Subject.Organization[0])

	require.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
	require.True(t, cert.BasicConstraintsValid)

	for _, h := range cert.DNSNames {
		require.Contains(t, args.hosts, h)
	}

	for _, h := range cert.IPAddresses {
		require.Contains(t, args.hosts, h.String())
	}

}

func requirePrivSignedCert(t *testing.T, priv any, cert *x509.Certificate) {
	t.Helper()

	type PK interface {
		Public() crypto.PublicKey
		Equal(x crypto.PrivateKey) bool
	}

	type PublicK interface {
		Equal(x crypto.PublicKey) bool
	}

	pk, ok := priv.(PK)
	require.True(t, ok)
	public, ok := pk.Public().(PublicK)
	require.True(t, ok)
	require.True(t, public.Equal(cert.PublicKey))
}

func newPrivKey(t *testing.T) any {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)
	return priv
}

func newRSAPrivKey(t *testing.T) any {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return priv
}

type generateCertArgs struct {
	priv       interface{}
	randSource io.Reader
	hosts      []string
	validFrom  time.Time
	validFor   time.Duration
	isCA       bool
}

func Test_generateCert(t *testing.T) {
	tests := []struct {
		name    string
		args    generateCertArgs
		wantErr bool
	}{
		{
			name: "success",
			args: generateCertArgs{
				priv:      newPrivKey(t),
				hosts:     []string{"localhost"},
				validFrom: time.Now(),
				validFor:  time.Hour,
			},
		},
		{
			name: "rsa key",
			args: generateCertArgs{
				priv:  newRSAPrivKey(t),
				hosts: []string{"localhost"},
			},
		},
		{
			name: "validFrom IsZero",
			args: generateCertArgs{
				priv:      newPrivKey(t),
				hosts:     []string{"localhost"},
				validFrom: time.Time{},
			},
		},
		{
			name: "validFrom is set",
			args: generateCertArgs{
				priv:      newPrivKey(t),
				hosts:     []string{"localhost"},
				validFrom: time.Now(),
			},
		},
		{
			name: "validFor is zero",
			args: generateCertArgs{
				priv:     newPrivKey(t),
				hosts:    []string{"localhost"},
				validFor: 0,
			},
		},
		{
			name: "validFor is non-zero",
			args: generateCertArgs{
				priv:     newPrivKey(t),
				hosts:    []string{"localhost"},
				validFor: 24 * time.Hour,
			},
		},
		{
			name: "isCA true",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"localhost"},
				isCA:  true,
			},
		},
		{
			name: "isCA false",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"localhost"},
				isCA:  false,
			},
		},
		{
			name: "DNS hostname",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"localhost"},
			},
		},
		{
			name: "IP hostname",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"127.0.0.1"},
			},
		},
		{
			name: "bad IP hostname",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"999.999.999.999"},
			},
		},
		{
			name: "Multiple mixed hostnames",
			args: generateCertArgs{
				priv:  newPrivKey(t),
				hosts: []string{"localhost", "127.0.0.1"},
			},
		},
		{
			name: "error generate serial",
			args: generateCertArgs{
				priv:       newPrivKey(t),
				randSource: iotest.ErrReader(assert.AnError),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateCert(tt.args.priv, tt.args.hosts, tt.args.validFrom, tt.args.validFor, tt.args.isCA, tt.args.randSource)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			gotCert, err := x509.ParseCertificate(got)
			require.NoError(t, err)
			require.NotNil(t, gotCert)
			checkCert(t, gotCert, tt.args)
		})
	}
}

func Test_signCert(t *testing.T) {
	testKey := newPrivKey(t)
	type args struct {
		randSource io.Reader
		template   *x509.Certificate
		parent     *x509.Certificate
		pub        any
		priv       any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "return error",
			args: args{
				priv:     testKey,
				pub:      publicKey(testKey),
				template: &x509.Certificate{},
				parent:   &x509.Certificate{},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := signCert(tt.args.template, tt.args.parent, tt.args.pub, tt.args.priv)
			if !tt.wantErr(t, err, fmt.Sprintf("signCert(%v, %v, %v, %v, %v)", tt.args.randSource, tt.args.template, tt.args.parent, tt.args.pub, tt.args.priv)) {
				return
			}
			assert.Equalf(t, tt.want, got, "signCert(%v, %v, %v, %v, %v)", tt.args.randSource, tt.args.template, tt.args.parent, tt.args.pub, tt.args.priv)
		})
	}
}

func Test_ensureRand(t *testing.T) {
	type args struct {
		randSource io.Reader
	}
	tests := []struct {
		name string
		args args
		want io.Reader
	}{
		{
			name: "nil gives rand.Reader",
			args: args{
				randSource: nil,
			},
			want: rand.Reader,
		},
		{
			name: "returns provided io.Reader",
			args: args{
				iotest.ErrReader(assert.AnError),
			},
			want: iotest.ErrReader(assert.AnError),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ensureRand(tt.args.randSource), "ensureRand(%v)", tt.args.randSource)
		})
	}
}

func Test_generateKey(t *testing.T) {
	type args struct {
		keytype    KeyType
		randSource io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "unsupported key type",
			args:    args{},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "RSA4096",
			args: args{
				keytype: KeyTypeRSA4096,
			},
			want:    &rsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "ED25519",
			args: args{
				keytype: KeyTypeED25519,
			},
			want:    ed25519.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "ECDSA224",
			args: args{
				keytype: KeyTypeECP224,
			},
			want:    &ecdsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "ECDSA256",
			args: args{
				keytype: KeyTypeECP256,
			},
			want:    &ecdsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "ECDSA384",
			args: args{
				keytype: KeyTypeECP384,
			},
			want:    &ecdsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "ECDSA521",
			args: args{
				keytype: KeyTypeECP521,
			},
			want:    &ecdsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "returns error",
			args: args{
				keytype:    KeyTypeECP256,
				randSource: iotest.ErrReader(assert.AnError),
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateKey(tt.args.keytype, tt.args.randSource)
			if !tt.wantErr(t, err, fmt.Sprintf("generateKey(%v, %v)", tt.args.keytype, tt.args.randSource)) {
				return
			}
			assert.IsTypef(t, tt.want, got, "generateKey(%v, %v)", tt.args.keytype, tt.args.randSource)
		})
	}
}

func Test_newDefaultTLSConfig(t *testing.T) {
	tests := []struct {
		name string
		want *tls.Config
	}{
		{
			name: "default TLS config",
			want: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, newDefaultTLSConfig(), "newDefaultTLSConfig()")
		})
	}
}

func newED25519PrivKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return priv
}

func Test_publicKey(t *testing.T) {
	type args struct {
		priv any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "rsa",
			args: args{
				priv: newRSAPrivKey(t),
			},
			want: &rsa.PublicKey{},
		},
		{
			name: "ecdsa",
			args: args{
				priv: newPrivKey(t),
			},
			want: &ecdsa.PublicKey{},
		},
		{
			name: "ed25519",
			args: args{
				priv: newED25519PrivKey(t),
			},
			want: ed25519.PublicKey{},
		},
		{
			name: "nil",
			args: args{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsTypef(t, tt.want, publicKey(tt.args.priv), "publicKey(%v)", tt.args.priv)
		})
	}
}

func Test_readCertificate(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				path: "testdata/certKeyBundle_1/cert.pem",
			},
			want:    []byte{0x30, 0x82, 0x1, 0xe0, 0x30, 0x82, 0x1, 0x85, 0xa0, 0x3, 0x2, 0x1, 0x2, 0x2, 0x14, 0x7f, 0x31, 0xec, 0xdf, 0x7a, 0x43, 0x56, 0x8d, 0x36, 0x64, 0x3e, 0x4b, 0x49, 0xdc, 0xfd, 0xf5, 0xa8, 0x64, 0x8f, 0xf, 0x30, 0xa, 0x6, 0x8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x2, 0x30, 0x45, 0x31, 0xb, 0x30, 0x9, 0x6, 0x3, 0x55, 0x4, 0x6, 0x13, 0x2, 0x41, 0x55, 0x31, 0x13, 0x30, 0x11, 0x6, 0x3, 0x55, 0x4, 0x8, 0xc, 0xa, 0x53, 0x6f, 0x6d, 0x65, 0x2d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x31, 0x21, 0x30, 0x1f, 0x6, 0x3, 0x55, 0x4, 0xa, 0xc, 0x18, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x20, 0x57, 0x69, 0x64, 0x67, 0x69, 0x74, 0x73, 0x20, 0x50, 0x74, 0x79, 0x20, 0x4c, 0x74, 0x64, 0x30, 0x1e, 0x17, 0xd, 0x32, 0x34, 0x30, 0x35, 0x31, 0x32, 0x31, 0x37, 0x35, 0x39, 0x34, 0x30, 0x5a, 0x17, 0xd, 0x33, 0x34, 0x30, 0x35, 0x31, 0x30, 0x31, 0x37, 0x35, 0x39, 0x34, 0x30, 0x5a, 0x30, 0x45, 0x31, 0xb, 0x30, 0x9, 0x6, 0x3, 0x55, 0x4, 0x6, 0x13, 0x2, 0x41, 0x55, 0x31, 0x13, 0x30, 0x11, 0x6, 0x3, 0x55, 0x4, 0x8, 0xc, 0xa, 0x53, 0x6f, 0x6d, 0x65, 0x2d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x31, 0x21, 0x30, 0x1f, 0x6, 0x3, 0x55, 0x4, 0xa, 0xc, 0x18, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x20, 0x57, 0x69, 0x64, 0x67, 0x69, 0x74, 0x73, 0x20, 0x50, 0x74, 0x79, 0x20, 0x4c, 0x74, 0x64, 0x30, 0x59, 0x30, 0x13, 0x6, 0x7, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x2, 0x1, 0x6, 0x8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x3, 0x1, 0x7, 0x3, 0x42, 0x0, 0x4, 0x8f, 0x3b, 0xdf, 0x57, 0x9a, 0x98, 0x1e, 0x96, 0x81, 0x65, 0x56, 0x16, 0x80, 0xbb, 0xff, 0xcc, 0x70, 0xda, 0x66, 0xb7, 0x4a, 0x67, 0xef, 0x4e, 0x46, 0x5b, 0xcf, 0x62, 0x10, 0xd9, 0x6e, 0x98, 0x99, 0x56, 0x20, 0xeb, 0xec, 0xa6, 0xac, 0x3d, 0xb4, 0xc8, 0x2e, 0x31, 0x2b, 0xc2, 0xa4, 0x45, 0xc, 0xab, 0x2, 0x1, 0xb8, 0x49, 0x59, 0x6f, 0x67, 0x57, 0x8b, 0xb4, 0x27, 0x24, 0xf7, 0x6, 0xa3, 0x53, 0x30, 0x51, 0x30, 0x1d, 0x6, 0x3, 0x55, 0x1d, 0xe, 0x4, 0x16, 0x4, 0x14, 0x18, 0x54, 0xe1, 0x93, 0x17, 0x77, 0xe2, 0x75, 0xc1, 0xe9, 0xdd, 0xa, 0xf7, 0xdf, 0x4b, 0x64, 0x2d, 0x27, 0x14, 0x3a, 0x30, 0x1f, 0x6, 0x3, 0x55, 0x1d, 0x23, 0x4, 0x18, 0x30, 0x16, 0x80, 0x14, 0x18, 0x54, 0xe1, 0x93, 0x17, 0x77, 0xe2, 0x75, 0xc1, 0xe9, 0xdd, 0xa, 0xf7, 0xdf, 0x4b, 0x64, 0x2d, 0x27, 0x14, 0x3a, 0x30, 0xf, 0x6, 0x3, 0x55, 0x1d, 0x13, 0x1, 0x1, 0xff, 0x4, 0x5, 0x30, 0x3, 0x1, 0x1, 0xff, 0x30, 0xa, 0x6, 0x8, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x4, 0x3, 0x2, 0x3, 0x49, 0x0, 0x30, 0x46, 0x2, 0x21, 0x0, 0xd6, 0xd9, 0x1c, 0xec, 0xd8, 0xca, 0xd6, 0x96, 0x7b, 0xea, 0x41, 0xbe, 0xdb, 0xc4, 0xcc, 0x24, 0xac, 0x83, 0x98, 0x7c, 0xe1, 0x8d, 0x72, 0x2c, 0x32, 0xdd, 0x42, 0xd1, 0x4d, 0x4b, 0x1e, 0x77, 0x2, 0x21, 0x0, 0xa5, 0x6d, 0x7c, 0xbb, 0xfd, 0x71, 0xb7, 0xea, 0x61, 0xa8, 0x30, 0xd6, 0x95, 0xdc, 0x85, 0x94, 0x9d, 0x62, 0xe3, 0xa1, 0x7, 0xb6, 0xfa, 0xe6, 0x47, 0x52, 0x90, 0x4f, 0x6d, 0x27, 0xaa, 0x7d},
			wantErr: assert.NoError,
		},
		{
			name: "read key",
			args: args{
				path: "testdata/certKeyBundle_1/ec_key.pem",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readCertificate(tt.args.path)
			if !tt.wantErr(t, err, fmt.Sprintf("readCertificate(%v)", tt.args.path)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readCertificate(%v)", tt.args.path)
		})
	}
}

func Test_readPEMFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *pem.Block
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				path: "testdata/certKeyBundle_1/ec_key.pem",
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "file does not exist",
			args: args{
				path: "/doesnotexist_random_32049r8",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPEMFile(tt.args.path)
			if !tt.wantErr(t, err, fmt.Sprintf("readPEMFile(%v)", tt.args.path)) {
				return
			}
			assert.IsTypef(t, tt.want, got, "readPEMFile(%v)", tt.args.path)
		})
	}
}

func Test_readPrivateKeyPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ec private key",
			args: args{
				path: "testdata/certKeyBundle_1/ec_key.pem",
			},
			want:    &ecdsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "private key",
			args: args{
				path: "testdata/rsa2048_key.pem",
			},
			want:    &rsa.PrivateKey{},
			wantErr: assert.NoError,
		},
		{
			name: "invalid key",
			args: args{
				path: "testdata/invalid_key.pem",
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "not a private key",
			args: args{
				path: "testdata/certKeyBundle_1/cert.pem",
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "not PEM",
			args: args{
				path: "/doesnotexist_something_random_4309572",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPrivateKeyPath(tt.args.path)
			if !tt.wantErr(t, err, fmt.Sprintf("readPrivateKeyPath(%v)", tt.args.path)) {
				return
			}
			assert.IsTypef(t, tt.want, got, "readPrivateKeyPath(%v)", tt.args.path)
		})
	}
}
