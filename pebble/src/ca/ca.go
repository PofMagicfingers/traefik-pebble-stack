package ca

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"
	"io/ioutil"

	"github.com/zimosworld/pebble/acme"
	"github.com/zimosworld/pebble/core"
	"github.com/zimosworld/pebble/db"
)

const (
	rootCAPrefix         = "Pebble Root CA "
	intermediateCAPrefix = "Pebble Intermediate CA "
)

type CAImpl struct {
	log *log.Logger
	db  *db.MemoryStore

	root         *issuer
	intermediate *issuer
}

type issuer struct {
	key  crypto.Signer
	cert *core.Certificate
}

func makeSerial() *big.Int {
	serial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(fmt.Sprintf("unable to create random serial number: %s", err.Error()))
	}
	return serial
}

// makeKey and makeRootCert are adapted from MiniCA:
// https://github.com/jsha/minica/blob/3a621c05b61fa1c24bcb42fbde4b261db504a74f/main.go

// makeKey creates a new 2048 bit RSA private key
func makeKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (ca *CAImpl) getKey(keyFile string) (*rsa.PrivateKey, error) {

    kf, e := ioutil.ReadFile(keyFile)
    if e != nil {
        ca.log.Printf("kfload:", e.Error())
        return nil, e
    }

    kpb, kr := pem.Decode(kf)
    fmt.Println(string(kr))

    key, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
    if e != nil {
        ca.log.Printf("parsekey:", e.Error())
        return nil, e
    }

    return key, nil
}

func (ca *CAImpl) makeRootCert(
	subjectKey crypto.Signer,
	subjCNPrefix string,
	signer *issuer) (*core.Certificate, error) {

	serial := makeSerial()
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: subjCNPrefix + hex.EncodeToString(serial.Bytes()[:3]),
		},
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(30, 0, 0),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA: true,
	}

	var signerKey crypto.Signer
	if signer != nil && signer.key != nil {
		signerKey = signer.key
	} else {
		signerKey = subjectKey
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, subjectKey.Public(), signerKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	hexSerial := hex.EncodeToString(cert.SerialNumber.Bytes())
	newCert := &core.Certificate{
		ID:   hexSerial,
		Cert: cert,
		DER:  der,
	}
	if signer != nil && signer.cert != nil {
		newCert.Issuer = signer.cert
	}
	_, err = ca.db.AddCertificate(newCert)
	if err != nil {
		return nil, err
	}
	return newCert, nil
}

func (ca *CAImpl) getRootCert(certFile string) (*core.Certificate, error) {
    cf, e := ioutil.ReadFile(certFile)
    if e != nil {
        ca.log.Printf("cfload:", e.Error())
        return nil, e
    }

    cpb, cr := pem.Decode(cf)
    fmt.Println(string(cr))

    crt, e := x509.ParseCertificate(cpb.Bytes)
    if e != nil {
        ca.log.Printf("parsex509:", e.Error())
        return nil, e
    }

    hexSerial := hex.EncodeToString(crt.SerialNumber.Bytes())
    cert := &core.Certificate{
        ID:   hexSerial,
        Cert: crt,
        DER:  cf,
    }

    return cert, nil
}

func (ca *CAImpl) LoadX509KeyPair(certFile, keyFile string) (*core.Certificate, *rsa.PrivateKey, error) {
    cf, e := ioutil.ReadFile(certFile)
    if e != nil {
        ca.log.Printf("cfload:", e.Error())
        return nil, nil, e
    }

    kf, e := ioutil.ReadFile(keyFile)
    if e != nil {
        ca.log.Printf("kfload:", e.Error())
        return nil, nil, e
    }
    cpb, cr := pem.Decode(cf)
    fmt.Println(string(cr))
    kpb, kr := pem.Decode(kf)
    fmt.Println(string(kr))

    crt, e := x509.ParseCertificate(cpb.Bytes)
    if e != nil {
        ca.log.Printf("parsex509:", e.Error())
        return nil, nil, e
    }

    hexSerial := hex.EncodeToString(crt.SerialNumber.Bytes())
	cert := &core.Certificate{
		ID:   hexSerial,
		Cert: crt,
		DER:  cf,
	}

    key, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
    if e != nil {
        ca.log.Printf("parsekey:", e.Error())
        return nil, nil, e
    }
    return cert, key, nil
}

func (ca *CAImpl) getIssuer() error {

	// Make an intermediate private key
	ik, err := ca.getKey("/var/pebble/certs/ca/key.pem")
	if err != nil {
		return err
	}

	// Make an intermediate certificate with the root issuer
	ic, err := ca.getRootCert("/var/pebble/certs/ca/cert.pem")
	if err != nil {
		return err
	}
	ca.intermediate = &issuer{
		key:  ik,
		cert: ic,
	}
	ca.log.Printf("Generated new intermediate issuer with serial %s\n", ic.ID)
	return nil
}

func (ca *CAImpl) newCertificate(domains []string, key crypto.PublicKey) (*core.Certificate, error) {
	var cn string
	if len(domains) > 0 {
		cn = domains[0]
	} else {
		return nil, fmt.Errorf("must specify at least one domain name")
	}

	issuer := ca.intermediate
	if issuer == nil || issuer.cert == nil {
		return nil, fmt.Errorf("cannot sign certificate - nil issuer")
	}

	serial := makeSerial()
	template := &x509.Certificate{
		DNSNames: domains,
		Subject: pkix.Name{
			CommonName: cn,
		},
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA: false,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, issuer.cert.Cert, key, issuer.key)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	hexSerial := hex.EncodeToString(cert.SerialNumber.Bytes())
	newCert := &core.Certificate{
		ID:     hexSerial,
		Cert:   cert,
		DER:    der,
		Issuer: issuer.cert,
	}
	_, err = ca.db.AddCertificate(newCert)
	if err != nil {
		return nil, err
	}
	return newCert, nil
}

func New(log *log.Logger, db *db.MemoryStore) *CAImpl {
	ca := &CAImpl{
		log: log,
		db:  db,
	}
	err := ca.getIssuer()
	if err != nil {
		panic(fmt.Sprintf("Error creating new intermediate issuer: %s", err.Error()))
	}
	return ca
}

func (ca *CAImpl) CompleteOrder(order *core.Order) {
	// Lock the order for writing
	order.Lock()
	// If the order isn't pending, produce an error and immediately unlock
	if order.Status != acme.StatusPending {
		ca.log.Printf("Error: Asked to complete order %s is not status pending, was status %s",
			order.ID, order.Status)
		order.Unlock()
		return
	}
	// Otherwise update the order to be in a processing state
	order.Status = acme.StatusProcessing
	// Unlock the order again
	order.Unlock()

	// Check the authorizations - this is done by the VA before calling
	// CompleteOrder but we do it again for robustness sake.
	for _, authz := range order.AuthorizationObjects {
		// Lock the authorization for reading
		authz.RLock()
		if authz.Status != acme.StatusValid {
			return
		}
		authz.RUnlock()
	}

	// issue a certificate for the csr
	csr := order.ParsedCSR
	cert, err := ca.newCertificate(csr.DNSNames, csr.PublicKey)
	if err != nil {
		ca.log.Printf("Error: unable to issue order: %s", err.Error())
		return
	}
	ca.log.Printf("Issued certificate serial %s for order %s\n", cert.ID, order.ID)

	// Lock and update the order to valid status and store a cert ID for the wfe
	// to use to render the certificate URL for the order
	order.Lock()
	order.Status = acme.StatusValid
	order.CertificateObject = cert
	order.Unlock()
}
