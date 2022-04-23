package jwt

import (
	"crypto/rsa"
	"math/rand"
	"testing"
)

func TestRS256(t *testing.T) {
	reader := rand.New(rand.NewSource(1))
	private, err := rsa.GenerateKey(reader, 2048)
	if err != nil {
		t.Errorf("expecting error nil but got %v", err)
	}

	const content = "content-example"
	kid := "kid-example"
	signer := NewRS265Signer(kid, private)

	signature, err := signer.Sign([]byte(content))
	if err != nil {
		t.Errorf("expecting error nil but got %v", err)
	}

	verifier := NewRS256Verifier(&private.PublicKey)

	err = verifier.Verify([]byte(content), signature)
	if err != nil {
		t.Errorf("expecting error nil but got %v", err)
	}

	err = verifier.Verify([]byte("invalid content"), signature)
	if err == nil {
		t.Errorf("expecting error non-nil")
	}
}
