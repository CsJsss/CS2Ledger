package eco

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
)

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	// Allow literal \n as a portable substitute for actual newlines.
	pemStr = strings.ReplaceAll(pemStr, "\\n", "\n")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA private key")
		}
		return rsaKey, nil
	}

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key (PKCS8: %v, PKCS1: %v)", err, err)
	}
	return rsaKey, nil
}

func generateRSASignature(privateKey *rsa.PrivateKey, params map[string]any) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})

	parts := make([]string, 0, len(params))
	for _, k := range keys {
		v := params[k]
		if v == nil {
			continue
		}
		var vs string
		switch val := v.(type) {
		case map[string]any, []any:
			b, err := marshalCompact(val)
			if err != nil {
				return "", fmt.Errorf("marshal param %s: %w", k, err)
			}
			vs = string(b)
		default:
			vs = fmt.Sprintf("%v", val)
		}
		parts = append(parts, k+"="+vs)
	}
	message := strings.Join(parts, "&")

	hashed := sha256.Sum256([]byte(message))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("sign failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func marshalCompact(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	return b, nil
}
