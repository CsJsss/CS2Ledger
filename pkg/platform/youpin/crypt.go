package youpin

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

const uuPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv9BDdhCDahZNFuJeesx3gzoQfD7pE0AeWiNBZlc21ph6kU9zd58X/1warV3C1VIX0vMAmhOcj5u86i+L2Lb2V68dX2Nb70MIDeW6Ibe8d0nF8D30tPsM7kaAyvxkY6ECM6RHGNhV4RrzkHmf5DeR9bybQGE0A9jcjuxszD1wsW/n19eeom7MroHqlRorp5LLNR8bSbmhTw6M/RQ/Fm3lKjKcvs1QNVyBNimrbD+ZVPE/KHSZLQ1jdF6tppvFnGxgJU9NFmxGFU0hx6cZiQHkhOQfGDFkElxgtj8gFJ1narTwYbvfe5nGSiznv/EUJSjTHxzX1TEkex0+5j4vSANt1QIDAQAB
-----END PUBLIC KEY-----`

// uuApiCrypt mirrors the Python UUApiCrypt class for encrypting/decrypting
// communication with the YouPin deviceW2 endpoint.
type uuApiCrypt struct {
	aesKey    []byte
	publicKey *rsa.PublicKey
}

func newUUApiCrypt(aesKey string) (*uuApiCrypt, error) {
	block, _ := pem.Decode([]byte(uuPublicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("youpin crypt: failed to parse public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("youpin crypt: parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("youpin crypt: not an RSA public key")
	}
	return &uuApiCrypt{
		aesKey:    []byte(aesKey),
		publicKey: rsaPub,
	}, nil
}

// getEncryptedAesKey encrypts the AES key with the RSA public key (PKCS1v1.5)
// and returns the base64-encoded result.
func (c *uuApiCrypt) getEncryptedAesKey() (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, c.publicKey, c.aesKey)
	if err != nil {
		return "", fmt.Errorf("RSA encrypt AES key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// uuEncrypt encrypts content with AES-ECB (PKCS7 padding) and returns
// the base64-encoded result.
func (c *uuApiCrypt) uuEncrypt(content string) (string, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("AES new cipher: %w", err)
	}
	plaintext := []byte(content)
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	encrypted := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// uuDecrypt decrypts a base64-encoded AES-ECB ciphertext and returns
// the plaintext string.
func (c *uuApiCrypt) uuDecrypt(encryptedBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", fmt.Errorf("AES new cipher: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", fmt.Errorf("pkcs7 unpad: %w", err)
	}
	return string(plaintext), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	pad := make([]byte, padLen)
	for i := range pad {
		pad[i] = byte(padLen)
	}
	return append(data, pad...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padded data length %d", len(data))
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, fmt.Errorf("invalid padding length %d", padLen)
	}
	return data[:len(data)-padLen], nil
}

func randomUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
