package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tam7t/hpkp"
)

func main() {
	var err error

	serverPtr := flag.String("server", "", "server to inspect (ex: github.com:443)")
	filePtr := flag.String("file", "", "path to PEM encoded certificate")

	flag.Parse()

	if *filePtr != "" {
		err = fromFile(*filePtr)
	}

	if err != nil {
		log.Fatal(err)
	}

	if *serverPtr != "" {
		err = fromServer(*serverPtr)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func fromServer(server string) error {
	conn, err := tls.Dial("tcp", server, &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		return err
	}

	for _, cert := range conn.ConnectionState().PeerCertificates {
		fmt.Println(hpkp.Fingerprint(cert))
	}

	return nil
}

func fromFile(path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var block *pem.Block

	for len(contents) > 0 {
		block, contents = pem.Decode(contents)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(hpkp.Fingerprint(cert))
	}

	return nil
}
