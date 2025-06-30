package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/cert-manager/cert-manager/pkg/util/pki"
)

var linkerdCerts = []string{
	"linkerd/ca.crt",
	"linkerd/ca.key",
	"linkerd/issuer.crt",
	"linkerd/issuer.key",
}

var selfsigned = []string{
	"server.key",
	"server.crt",
}

func main() {
	genLinkerdCerts()
	genSelfSignerCert()
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func genSelfSignerCert() {

	dir := ""

	create, err := isCorrectCert("self-signed", selfsigned)
	if err != nil {
		log.Fatalln(err)
	}

	if create {

		hn, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("creating self-signed certs")

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 16384)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Fatal(err)
		}

		// set up our server certificate
		cert := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName: hn,
			},
			NotBefore:   time.Now().Add(-10 * time.Second),
			NotAfter:    time.Now().AddDate(10, 0, 0),
			KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}

		cert.DNSNames = []string{hn}

		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatal(err)
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, publicKey(priv), priv)
		if err != nil {
			log.Fatalf("Failed to create certificate: %s", err)
		}
		out := &bytes.Buffer{}
		pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		err = os.WriteFile(filepath.Join(dir, "server.crt"), out.Bytes(), 0400)
		if err != nil {
			log.Fatal(err)
		}

		out.Reset()
		pem.Encode(out, pemBlockForKey(priv))
		err = os.WriteFile(filepath.Join(dir, "server.key"), out.Bytes(), 0400)
		if err != nil {
			log.Fatal(err)
		}

	}
}

func isCorrectCert(name string, files []string) (bool, error) {

	exists := 0

	log.Printf("checking %s certificates\n", name)

	for a := range files {
		cert := files[a]
		if fileExists(cert) {
			exists++
		}
	}

	log.Printf("found %d certificates for %s\n", exists, name)

	if exists > 0 &&
		exists != len(files) {
		return false, fmt.Errorf("certificates for %s are out of sync. found %d certificates but should be %d.",
			name, exists, len(files))
	} else if exists > 0 && exists == len(files) {
		log.Printf("%s certificates exist", name)
		return false, nil
	}

	return true, nil

}

func genLinkerdCerts() {

	create, err := isCorrectCert("linkerd", linkerdCerts)
	if err != nil {
		log.Fatalln(err)
	}

	if create {
		// now we can create
		log.Println("create linkerd certificates")

		cert, key := createCA("linkerd")
		createIntermediateCA("linkerd", cert, key)
	}

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func createCA(dir string) (*x509.Certificate, *ecdsa.PrivateKey) {

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "root.linkerd.cluster.local",
		},
		Issuer: pkix.Name{
			CommonName: "root.linkerd.cluster.local",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	caPrivKey, err := pki.GenerateECPrivateKey(pki.ECCurve256)
	if err != nil {
		log.Fatal(err)
	}

	caPublicKey, err := pki.PublicKeyForPrivateKey(caPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, caPublicKey, caPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	encCa, err := pki.EncodeECPrivateKey(caPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "ca.crt"), caPEM.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dir, "ca.key"), encCa, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return ca, caPrivKey

}

func createIntermediateCA(dir string, ca *x509.Certificate, key *ecdsa.PrivateKey) {

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "identity.linkerd.cluster.local",
		},
		Issuer: pkix.Name{
			CommonName: "root.linkerd.cluster.local",
		},
		NotBefore:             time.Now().Add(-10 * time.Second),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
		MaxPathLen:            0,
		Extensions:            []pkix.Extension{},
	}

	certPrivKey, err := pki.GenerateECPrivateKey(pki.ECCurve256)
	if err != nil {
		log.Fatal(err)
	}

	certPublicKey, err := pki.PublicKeyForPrivateKey(certPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, certPublicKey, key)
	if err != nil {
		log.Fatal(err)
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	encCert, err := pki.EncodeECPrivateKey(certPrivKey)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(dir, "issuer.crt"), certPEM.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dir, "issuer.key"), encCert, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
