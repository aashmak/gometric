package crypto

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"io"

	"os"
)

func NewPublicKey(file string) (*rsa.PublicKey, error) {
	data, err := readFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey.(*rsa.PublicKey), err
}

func NewPrivateKey(file string) (*rsa.PrivateKey, error) {
	data, err := readFile(file)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey.(*rsa.PrivateKey), err
}

func readFile(file string) ([]byte, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0777)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	data := make([]byte, 0)

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return []byte{}, err
			}
		}

		data = append(data, line...)
	}

	return data, nil
}

func Encrypt(pubKey *rsa.PublicKey, data *bytes.Buffer) ([]byte, error) {
	encryptData, err := rsa.EncryptOAEP(
		sha512.New(),
		rand.Reader,
		pubKey,
		data.Bytes(),
		nil,
	)
	if err != nil {
		return []byte{}, err
	}

	return encryptData, err
}

func Decrypt(privKey *rsa.PrivateKey, data *bytes.Buffer) ([]byte, error) {
	decryptData, err := rsa.DecryptOAEP(
		sha512.New(),
		rand.Reader,
		privKey,
		data.Bytes(),
		nil,
	)

	return decryptData, err
}
