package tests

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/gokeyless/client"
	"github.com/cloudflare/gokeyless/server"
)

const (
	serverCert   = "testdata/server.pem"
	serverKey    = "testdata/server-key.pem"
	keylessCA    = "testdata/ca.pem"
	serverAddr   = "localhost:3407"
	rsaPrivKey   = "testdata/rsa.key"
	ecdsaPrivKey = "testdata/ecdsa.key"

	clientCert  = "testdata/client.pem"
	clientKey   = "testdata/client-key.pem"
	keyserverCA = "testdata/ca.pem"
	rsaPubKey   = "testdata/rsa.pubkey"
	ecdsaPubKey = "testdata/ecdsa.pubkey"
)

var (
	s        *server.Server
	c        *client.Client
	rsaKey   *client.PrivateKey
	ecdsaKey *client.PrivateKey
	host     string
	port     int
	remote   client.Remote
)

// Set up compatible server and client for use by tests.
func init() {
	var err error
	var pemBytes []byte
	var p *pem.Block
	var priv crypto.Signer
	var pub crypto.PublicKey
	var portStr string

	log.Level = log.LevelFatal

	s, err = server.NewServerFromFile(serverCert, serverKey, keylessCA, serverAddr, "")
	if err != nil {
		log.Fatal(err)
	}

	if pemBytes, err = ioutil.ReadFile(rsaPrivKey); err != nil {
		log.Fatal(err)
	}
	p, _ = pem.Decode(pemBytes)
	if priv, err = x509.ParsePKCS1PrivateKey(p.Bytes); err != nil {
		log.Fatal(err)
	}
	if err = s.Keys.Add(nil, priv); err != nil {
		log.Fatal(err)
	}

	if pemBytes, err = ioutil.ReadFile(ecdsaPrivKey); err != nil {
		log.Fatal(err)
	}
	p, _ = pem.Decode(pemBytes)
	if priv, err = x509.ParseECPrivateKey(p.Bytes); err != nil {
		log.Fatal(err)
	}
	if err = s.Keys.Add(nil, priv); err != nil {
		log.Fatal(err)
	}

	listening := make(chan bool)
	go func() {
		listening <- true
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	<-listening

	if c, err = client.NewClientFromFile(clientCert, clientKey, keyserverCA); err != nil {
		log.Fatal(err)
	}

	if host, portStr, err = net.SplitHostPort(serverAddr); err != nil {
		log.Fatal(err)
	}
	if port, err = strconv.Atoi(portStr); err != nil {
		log.Fatal(err)
	}
	remote, err = c.LookupServer(host, "", port)
	if err != nil {
		log.Fatal(err)
	}

	if pemBytes, err = ioutil.ReadFile(rsaPubKey); err != nil {
		log.Fatal(err)
	}
	p, _ = pem.Decode(pemBytes)
	if pub, err = x509.ParsePKIXPublicKey(p.Bytes); err != nil {
		log.Fatal(err)
	}
	if rsaKey, err = c.RegisterPublicKey(serverAddr, pub); err != nil {
		log.Fatal(err)
	}

	if pemBytes, err = ioutil.ReadFile(ecdsaPubKey); err != nil {
		log.Fatal(err)
	}
	p, _ = pem.Decode(pemBytes)
	if pub, err = x509.ParsePKIXPublicKey(p.Bytes); err != nil {
		log.Fatal(err)
	}
	if ecdsaKey, err = c.RegisterPublicKey(serverAddr, pub); err != nil {
		log.Fatal(err)
	}
}
