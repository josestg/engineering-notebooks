package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

type RS265Signer struct {
	kid string
	alg string
	key *rsa.PrivateKey
}

func NewRS265Signer(kid string, key *rsa.PrivateKey) *RS265Signer {
	return &RS265Signer{
		kid: kid,
		key: key,
		alg: "RS256",
	}
}

func (r *RS265Signer) Alg() string {
	return r.alg
}

func (r *RS265Signer) KeyID() string {
	return r.kid
}

func (r *RS265Signer) Sign(b []byte) (sig []byte, err error) {
	hashFunc := crypto.SHA256.New()
	hashFunc.Write(b)
	hashed := hashFunc.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, r.key, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("%w: signing using RSA", err)
	}

	return signature, nil
}

type RS256Verifier struct {
	key *rsa.PublicKey
}

func NewRS256Verifier(key *rsa.PublicKey) *RS256Verifier {
	return &RS256Verifier{
		key: key,
	}
}

func (r *RS256Verifier) Verify(b []byte, sig []byte) error {
	hashFunc := crypto.SHA256.New()
	hashFunc.Write(b)
	hashed := hashFunc.Sum(nil)

	if err := rsa.VerifyPKCS1v15(r.key, crypto.SHA256, hashed[:], sig); err != nil {
		return fmt.Errorf("%w: verifying signature using RSA", err)
	}

	return nil
}
