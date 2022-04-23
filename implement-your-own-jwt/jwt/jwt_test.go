package jwt

import (
	"crypto/rsa"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestNewTime(t *testing.T) {

	t1 := NewTime(time.Now())
	t2 := new(Time)

	b, err := t1.MarshalJSON()
	if err != nil {
		t.Fatalf("expecting error nil but got %v", err)
	}

	if err := t2.UnmarshalJSON(b); err != nil {
		t.Fatalf("expecting error nil but got %v", err)
	}

	if !reflect.DeepEqual(t1, t2) {
		t.Fatalf("expecting t1 and t2 are equal")
	}

}

func TestStandardClaims_Verify(t *testing.T) {
	tests := []struct {
		desc       string
		claims     StandardClaims
		verifyTime *Time
		err        error
	}{
		{
			desc:       "empty claims should be always valid",
			claims:     StandardClaims{},
			verifyTime: NewTime(time.Now()),
			err:        nil,
		},
		{
			desc: "claims not activated yet",
			claims: StandardClaims{
				NotBefore: NewTime(time.Now().Add(time.Hour)),
			},
			verifyTime: NewTime(time.Now()),
			err:        ErrNotBefore,
		},
		{
			desc: "claims expired",
			claims: StandardClaims{
				ExpiresAt: NewTime(time.Now().Add(-time.Hour)),
			},
			verifyTime: NewTime(time.Now()),
			err:        ErrExpired,
		},
		{
			desc: "complete claims also valid",
			claims: StandardClaims{
				ID:        "123",
				Issuer:    "just for func",
				Subject:   "subject",
				Audience:  "service",
				IssuedAt:  NewTime(time.Now()),
				ExpiresAt: NewTime(time.Now().Add(time.Hour)),
				NotBefore: NewTime(time.Now()),
			},
			verifyTime: NewTime(time.Now()),
			err:        nil,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.desc, func(t *testing.T) {
			if err := tc.claims.Valid(tc.verifyTime); err != tc.err {
				t.Errorf("expecting %v but got %v", tc.err, err)
			}
		})
	}

}

func TestJWT_RS256(t *testing.T) {
	reader := rand.New(rand.NewSource(1))
	private, err := rsa.GenerateKey(reader, 2048)
	if err != nil {
		t.Fatalf("expecting nil but got %v", err)
	}

	keyID := "keyID-example"
	signer := NewRS265Signer(keyID, private)

	claims := StandardClaims{
		ID:        "123",
		Issuer:    "implement-your-own-jwt",
		Subject:   "12345",
		IssuedAt:  NewTime(time.Now()),
		ExpiresAt: NewTime(time.Now().Add(time.Hour)),
	}

	token, err := CreateToken(signer, Header{}, claims)
	if err != nil {
		t.Fatalf("expecting nil but got %v", err)
	}

	t.Log(token)

	selector := func(alg, kid string) (Verifier, error) {
		if alg == "RS256" && kid == keyID {
			return NewRS256Verifier(&private.PublicKey), nil
		}

		return nil, errors.New("unknown verifier")
	}

	var claims2 StandardClaims
	if err = DecodeToken(selector, token, &claims2); err != nil {
		t.Fatalf("expecting nil but got %v", err)
	}

	if !reflect.DeepEqual(claims, claims2) {
		t.Fatalf("expecting claims are equal")
	}
}
