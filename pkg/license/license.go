// Package license provides functionality for creating,
// signing, and validating licenses using a embedded public key
package license

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

//go:embed public-key.pem
var PublicKey string

// License represents the details of a software license, including customer
// information, type, features, expiry, and a digital signature.
type License struct {
	To        string   `json:"to"`
	Type      string   `json:"type"`
	Features  []string `json:"features"`
	ExpiresAt string   `json:"expiresAt"`
	Signature string   `json:"signature"`
}

// Sign creates a new license, signs it using the provided private key,
// and sets default values for fields if they are not provided.
// It returns an error if the private key is invalid or the signing fails.
func Sign(l License, key []byte) (string, error) {
	if l.ExpiresAt == "" {
		l.ExpiresAt = time.Now().AddDate(1, 0, 0).Format(time.RFC3339)
	}
	_, err := time.Parse(time.RFC3339, l.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("invalid License.ExpiresAt: '%s', err: %w", l.ExpiresAt, err)
	}

	// parse private key
	block, _ := pem.Decode(key)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	digest, err := marshalAndDigest(&l)
	if err != nil {
		return "", err
	}

	// generate signature
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify verifies the integrity, authenticity, and validity of a license
// by checking its expiry date and its digital signature using a pre-embedded public key.
// It returns an error if the license is expired or the signature is invalid.
func Verify(l License, pk []byte) error {
	expiresAt, err := time.Parse(time.RFC3339, l.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid License.ExpiresAt: '%s', err: %w", l.ExpiresAt, err)
	}
	// Check if the license has expired
	if time.Now().After(expiresAt) {
		return fmt.Errorf("license expired since %s", l.ExpiresAt)
	}

	signature := l.Signature

	// Clean signature for comparison of signature
	l.Signature = ""

	digest, err := marshalAndDigest(&l)
	if err != nil {
		return err
	}

	// Read public key
	block, _ := pem.Decode(pk)
	if block == nil {
		return fmt.Errorf("failed to decode public key")
	}

	pkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := pkey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("got unexpected key type: %T", pkey)
	}

	sign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify the digital signature
	return rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, digest, sign)
}

func VerifyJSON(str string, pk string) error {
	l := License{}
	err := json.Unmarshal([]byte(str), &l)
	if err != nil {
		return fmt.Errorf("invalid json format: %w", err)
	}

	return Verify(l, []byte(pk))
}

func marshalAndDigest(l *License) ([]byte, error) {
	// marshal license to binary
	lb, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	_, err = hash.Write(lb)
	if err != nil {
		return nil, err
	}
	digest := hash.Sum(nil)

	return digest, nil
}

func GenerateKeys() ([]byte, []byte, error) {
	// Step 2: Generate an RSA private key (PKCS #1 format)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generating rsa key: %w", err)
	}

	// Step 3: Create a self-signed X.509 certificate
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "direktiv.io",
			Organization: []string{"Direktiv"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	_, err = x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creating certificate: %w", err)
	}

	// Step 4: Save the private key in PKCS #1 format
	keyFile := bytes.NewBuffer(nil)
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		return nil, nil, fmt.Errorf("encoding private key: %w", err)
	}

	// Step 5: Store the public key in memory (PKCS #1 format)
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling public key: %w", err)
	}
	var publicKeyFile bytes.Buffer
	err = pem.Encode(&publicKeyFile, &pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})
	if err != nil {
		return nil, nil, fmt.Errorf("encoding public key: %w", err)
	}

	return keyFile.Bytes(), publicKeyFile.Bytes(), nil
}
