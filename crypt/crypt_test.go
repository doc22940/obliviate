package crypt

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"obliviate/config"
	"obliviate/crypt/rsa"
	"obliviate/interfaces/store/mock"
	"os"
	"testing"
	"time"
)

var conf *config.Configuration

func init() {
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	logrus.SetFormatter(formatter)
	//logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.FatalLevel)

	conf = &config.Configuration{
		DefaultDurationTime:     time.Hour * 24 * 7,
		ProdEnv:                 os.Getenv("ENV") == "PROD",
		MasterKey:               os.Getenv("HSM_MASTER_KEY"),
		KmsCredentialFile:       os.Getenv("KMS_CREDENTIAL_FILE"),
		FirestoreCredentialFile: os.Getenv("FIRESTORE_CREDENTIAL_FILE"),
	}
	//conf.Db = store.Connect(context.Background(), "test")
	conf.Db = mock.StorageMock()
}

func TestKeysGenerationAndStorage(t *testing.T) {

	rsa := rsa.NewMockAlgorithm()
	//rsa := rsa.NewAlgorithm()

	keys, err := NewKeys(conf, rsa)
	assert.NoError(t, err, "should not be error")

	pubKey := keys.PublicKeyEncoded

	var priv [32]byte
	var pub [32]byte
	pub = *keys.PublicKey
	priv = *keys.PrivateKey

	keys, err = NewKeys(conf, rsa)
	assert.NoError(t, err, "should not be error")

	assert.Equal(t, pubKey, keys.PublicKeyEncoded, "private keys should be the same")
	assert.Equal(t, priv, *keys.PrivateKey, "private keys should be the same")
	assert.Equal(t, pub, *keys.PublicKey, "public keys should be the same")

}
