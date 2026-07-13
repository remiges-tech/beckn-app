package catalog

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/blake2b"
)

// BuildAuthHeader builds a Beckn HTTP Signature Authorization header value.
//
// keyID must be in the format "<subscriberID>|<uniqueKeyID>|ed25519".
// privateKeyBase64 is a base64-encoded Ed25519 private key (64 bytes) or seed (32 bytes).
//
// Returns an empty string and no error when privateKeyBase64 is empty (signing skipped).
func BuildAuthHeader(body []byte, keyID, privateKeyBase64 string) (string, error) {
	return becknAuthHeader(body, keyID, privateKeyBase64)
}

func becknAuthHeader(body []byte, keyID, privateKeyBase64 string) (string, error) {
	if privateKeyBase64 == "" {
		return "", nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		keyBytes, err = base64.RawStdEncoding.DecodeString(privateKeyBase64)
		if err != nil {
			return "", fmt.Errorf("decode BPP private key: %w", err)
		}
	}

	var privKey ed25519.PrivateKey
	switch len(keyBytes) {
	case ed25519.PrivateKeySize:
		privKey = ed25519.PrivateKey(keyBytes)
	case ed25519.SeedSize:
		privKey = ed25519.NewKeyFromSeed(keyBytes)
	default:
		return "", fmt.Errorf("BPP private key length %d is invalid (expected 32 or 64 bytes)", len(keyBytes))
	}

	now := time.Now().Unix()
	expires := now + 600 // 10-minute window

	h, _ := blake2b.New512(nil)
	h.Write(body)
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	signingString := fmt.Sprintf("(created): %d\n(expires): %d\ndigest: BLAKE-512=%s",
		now, expires, digest)

	sig := ed25519.Sign(privKey, []byte(signingString))

	return fmt.Sprintf(
		`Signature keyId="%s",algorithm="ed25519",created="%d",expires="%d",headers="(created) (expires) digest",signature="%s"`,
		keyID, now, expires, base64.StdEncoding.EncodeToString(sig),
	), nil
}
