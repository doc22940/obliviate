package rsa

import (
	cloudkms "cloud.google.com/go/kms/apiv1"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"obliviate/config"
)

type RSA interface {
	EncryptRSA(*config.Configuration, []byte) ([]byte, error)
	DecryptRSA(*config.Configuration, []byte) ([]byte, error)
}

type Algorithm struct{}

func NewAlgorithm() Algorithm {
	return Algorithm{}
}

// encryptRSA will encrypt data locally using an 'RSA_DECRYPT_OAEP_2048_SHA256'
// public key retrieved from Cloud KMS.
func (Algorithm) EncryptRSA(conf *config.Configuration, plaintext []byte) ([]byte, error) {
	var err error
	var client *cloudkms.KeyManagementClient

	ctx := context.Background()
	if conf.ProdEnv {
		client, err = cloudkms.NewKeyManagementClient(ctx)
	} else {
		client, err = cloudkms.NewKeyManagementClient(ctx, option.WithCredentialsFile(conf.KmsCredentialFile))
	}
	if err != nil {
		return nil, fmt.Errorf("cloudkms.NewKeyManagementClient: %v", err)
	}

	// Retrieve the public key from KMS.
	response, err := client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{Name: conf.MasterKey})
	if err != nil {
		return nil, fmt.Errorf("client.GetPublicKey: %v", err)
	}

	// Parse the key.
	block, _ := pem.Decode([]byte(response.Pem))
	abstractKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509.ParsePKIXPublicKey: %+v", err)
	}

	rsaKey, ok := abstractKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key %v is not RSA", conf.MasterKey)
	}

	// Encrypt data using the RSA public key.
	cipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa.EncryptOAEP: %v", err)
	}

	return cipherText, nil
}

// decryptRSA will attempt to decrypt a given ciphertext with an
// private key stored on Cloud KMS.
func (Algorithm) DecryptRSA(conf *config.Configuration, ciphertext []byte) ([]byte, error) {
	var err error
	var client *cloudkms.KeyManagementClient

	ctx := context.Background()
	if conf.ProdEnv {
		client, err = cloudkms.NewKeyManagementClient(ctx)
	} else {
		client, err = cloudkms.NewKeyManagementClient(ctx, option.WithCredentialsFile(conf.KmsCredentialFile))
	}
	if err != nil {
		return nil, fmt.Errorf("cloudkms.NewKeyManagementClient: %v", err)
	}

	// Build the request.
	req := &kmspb.AsymmetricDecryptRequest{
		Name:       conf.MasterKey,
		Ciphertext: ciphertext,
	}
	// Call the API.
	response, err := client.AsymmetricDecrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AsymmetricDecrypt: %v", err)
	}
	return response.Plaintext, nil
}
