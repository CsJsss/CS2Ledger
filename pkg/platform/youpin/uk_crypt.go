package youpin

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
)

// ukPublicKey is the hardcoded RSA public key from YouPin's Android app.
const ukPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv9BDdhCDahZNFuJeesx3gzoQfD7pE0AeWiNBZlc21ph6kU9zd58X/1warV3C1VIX0vMAmhOcj5u86i+L2Lb2V68dX2Nb70MIDeW6Ibe8d0nF8D30tPsM7kaAyvxkY6ECM6RHGNhV4RrzkHmf5DeR9bybQGE0A9jcjuxszD1wsW/n19eeom7MroHqlRorp5LLNR8bSbmhTw6M/RQ/Fm3lKjKcvs1QNVyBNimrbD+ZVPE/KHSZLQ1jdF6tppvFnGxgJU9NFmxGFU0hx6cZiQHkhOQfGDFkElxgtj8gFJ1narTwYbvfe5nGSiznv/EUJSjTHxzX1TEkex0+5j4vSANt1QIDAQAB
-----END PUBLIC KEY-----`

var parsedUKPubKey *rsa.PublicKey

func init() {
	// Parse the hardcoded UK public key.
	block, _ := pem.Decode([]byte(ukPublicKey))
	if block == nil {
		panic("youpin: failed to decode hardcoded UK public key PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("youpin: failed to parse UK public key: %v", err))
	}
	var ok bool
	parsedUKPubKey, ok = key.(*rsa.PublicKey)
	if !ok {
		panic("youpin: UK public key is not RSA")
	}

	// Ensure crypto/sha1 is linked (needed for PKCS1v15 via crypto/rsa).
	_ = sha1.New
}

// ukCrypt implements YouPin's device registration encryption.
// AES-ECB for payload, RSA PKCS1v1.5 for wrapping the AES key.
type ukCrypt struct {
	aesKey []byte
}

func newUKCrypt() *ukCrypt {
	key := make([]byte, 16)
	_, _ = rand.Read(key)
	return &ukCrypt{aesKey: key}
}

// encryptedAESKey returns the RSA-encrypted, base64-encoded AES key.
func (c *ukCrypt) encryptedAESKey() (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, parsedUKPubKey, c.aesKey)
	if err != nil {
		return "", fmt.Errorf("uk: encrypt aes key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// encrypt encrypts content with AES-ECB and returns base64.
func (c *ukCrypt) encrypt(content string) (string, error) {
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}
	plaintext := []byte(content)
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:], plaintext[i:])
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt base64-decodes, AES-ECB decrypts, and PKCS7-unpads.
func (c *ukCrypt) decrypt(encryptedBase64 []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(encryptedBase64))
	if err != nil {
		return "", fmt.Errorf("uk: base64 decode: %w", err)
	}
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("uk: ciphertext not multiple of block size")
	}
	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:], ciphertext[i:])
	}
	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("uk: unpad: %w", err)
	}
	return string(unpadded), nil
}

// fetchUK calls /api/deviceW2 to get a fresh UK token.
func (c *Client) fetchUK() (string, error) {
	crypt := newUKCrypt()

	iud := randomUUID()
	payload := map[string]string{"iud": iud}
	payloadBytes, _ := json.Marshal(payload)

	encryptedData, err := crypt.encrypt(string(payloadBytes))
	if err != nil {
		return "", err
	}
	encryptedAESKey, err := crypt.encryptedAESKey()
	if err != nil {
		return "", err
	}

	body := map[string]string{
		"encryptedData":   encryptedData,
		"encryptedAesKey": encryptedAESKey,
	}
	bodyBytes, _ := json.Marshal(body)

	resp, err := c.HTTP.Post(c.BaseURL+"/api/deviceW2", "application/json; charset=utf-8", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("uk: deviceW2 request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("uk: deviceW2 returned %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decrypted, err := crypt.decrypt(respBody)
	if err != nil {
		return "", fmt.Errorf("uk: decrypt response: %w", err)
	}

	var ukResp struct {
		U string `json:"u"`
	}
	if err := json.Unmarshal([]byte(decrypted), &ukResp); err != nil {
		return "", fmt.Errorf("uk: parse response: %w", err)
	}

	return ukResp.U, nil
}

func randomUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("uk: empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, fmt.Errorf("uk: invalid padding")
	}
	return data[:len(data)-padding], nil
}
