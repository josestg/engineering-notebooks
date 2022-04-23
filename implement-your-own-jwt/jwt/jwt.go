package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

var (
	ErrExpired       = errors.New("jwt: claims expired")
	ErrNotBefore     = errors.New("jwt: claims not active yet")
	ErrInvalidFormat = errors.New("jwt: invalid token format")
)

// Time represents a JWT time.
// This Time truncates the time.Time with time.Millisecond.
//
// Time overrides the MarshalJSON and UnmarshalJSON of time.Time
// to make the marshaled time as plaintext number instead
// of formatted-string like time.RFC3339.
type Time struct {
	time.Time
}

// NewTime creates a new time at given time.
func NewTime(at time.Time) *Time {
	return &Time{at.Truncate(time.Millisecond)}
}

func (t *Time) MarshalJSON() ([]byte, error) {
	ms := t.Truncate(time.Millisecond).UnixMilli()
	return json.Marshal(ms)
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var ms json.Number
	if err := json.Unmarshal(b, &ms); err != nil {
		return err
	}

	msi64, err := ms.Int64()
	if err != nil {
		return err
	}

	mst := time.UnixMilli(msi64)
	*t = Time{mst}
	return nil
}

// Valid knows how validate claims.
type Valid interface {
	// Valid returns an error if the claims is not valid at the given time.
	Valid(at *Time) error
}

// StandardClaims is a structured version of Claims sections, as referenced at
// https://tools.ietf.org/html/rfc7519#section-4.1.
type StandardClaims struct {
	ID        string `json:"jti,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	IssuedAt  *Time  `json:"iat,omitempty"`
	ExpiresAt *Time  `json:"exp,omitempty"`
	NotBefore *Time  `json:"nbf,omitempty"`
}

func (s StandardClaims) Valid(at *Time) error {
	if s.ExpiresAt != nil && at.After(s.ExpiresAt.Time) {
		return ErrExpired
	}

	if s.NotBefore != nil && at.Before(s.NotBefore.Time) {
		return ErrNotBefore
	}

	return nil
}

// Header represents JWT header.
type Header map[string]interface{}

type Signer interface {
	Alg() string
	KeyID() string
	Sign(b []byte) (sig []byte, err error)
}

type Verifier interface {
	Verify(b []byte, sig []byte) error
}

type VerifierSelector func(alg string, kid string) (Verifier, error)

func CreateToken(signer Signer, header Header, claims interface{}) (string, error) {
	// injects signer info into header.
	header["alg"] = signer.Alg()
	header["kid"] = signer.KeyID()

	headerB64Encoded, err := b64URLEncoded(header)
	if err != nil {
		return "", fmt.Errorf("%w: encode header part", err)
	}

	log.Println(header)

	claimsB64Encoded, err := b64URLEncoded(claims)
	if err != nil {
		return "", fmt.Errorf("%w: encode claims part", err)
	}

	// JWT formats:
	// 	base64UrlEncode(header) + "." + base64UrlEncode(claims) + "." + base64UrlEncode(signature)
	sb := bytes.Buffer{}
	sb.Write(headerB64Encoded)
	sb.WriteByte('.')
	sb.Write(claimsB64Encoded)

	signature, err := signer.Sign(sb.Bytes())
	if err != nil {
		return "", fmt.Errorf("%w: create signature", err)
	}

	sb.WriteRune('.') // appends last '.' before the encoded signature.
	_, err = sb.WriteString(base64.RawURLEncoding.EncodeToString(signature))
	return sb.String(), err
}

func DecodeToken(selector VerifierSelector, token string, claims interface{}) error {
	// JWT formats:
	// 	base64UrlEncode(header) + "." + base64UrlEncode(payload) + "." + base64UrlEncode(signature)
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return ErrInvalidFormat
	}

	var header Header
	if err := b64URLEncodeToJSON(parts[0], &header); err != nil {
		return fmt.Errorf("%w: encode base64-url header into Header", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("%w: creating signature", err)
	}

	kid, _ := header["kid"].(string)
	alg, _ := header["alg"].(string)

	verifier, err := selector(alg, kid)
	if err != nil {
		return fmt.Errorf("%w: selecting verifier", err)
	}

	content := fmt.Sprintf("%s.%s", parts[0], parts[1])
	if err := verifier.Verify([]byte(content), signature); err != nil {
		return fmt.Errorf("%w: verifying signature", err)
	}

	if err := b64URLEncodeToJSON(parts[1], claims); err != nil {
		return fmt.Errorf("%w: encode base64-url body into claims", err)
	}

	if t, ok := claims.(Valid); ok {
		return t.Valid(NewTime(time.Now()))
	}

	return nil
}

func b64URLEncodeToJSON(urlEncoded string, v interface{}) error {
	reader := base64.NewDecoder(base64.RawURLEncoding, strings.NewReader(urlEncoded))
	return json.NewDecoder(reader).Decode(v)
}

func b64URLEncoded(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, base64.RawURLEncoding.EncodedLen(len(b)))
	base64.RawURLEncoding.Encode(buf, b)
	return buf, nil
}
