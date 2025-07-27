package ccrypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
)

const ChiselKeyPrefix = "ck-"

//  Relations between entities:
//
//   .............> PEM <...........
//   .               ^             .
//   .               |             .
//   .               |             .
//   .               V             .
//   ..........> ChiselKey .........

func privateKey2ChiselKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	encodedPrivateKey := make([]byte, base64.RawStdEncoding.EncodedLen(len(b)))
	base64.RawStdEncoding.Encode(encodedPrivateKey, b)

	return append([]byte(ChiselKeyPrefix), encodedPrivateKey...), nil
}

func privateKey2PEM(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
}

func chiselKey2PrivateKey(chiselKey []byte) (*ecdsa.PrivateKey, error) {
	rawChiselKey := chiselKey[len(ChiselKeyPrefix):]

	decodedPrivateKey := make([]byte, base64.RawStdEncoding.DecodedLen(len(rawChiselKey)))
	_, err := base64.RawStdEncoding.Decode(decodedPrivateKey, rawChiselKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseECPrivateKey(decodedPrivateKey)
}

func ChiselKey2PEM(chiselKey []byte) ([]byte, error) {
	privateKey, err := chiselKey2PrivateKey(chiselKey)
	if err == nil {
		return privateKey2PEM(privateKey)
	}

	return nil, err
}

func IsChiselKey(chiselKey []byte) bool {
	return strings.HasPrefix(string(chiselKey), ChiselKeyPrefix)
}
