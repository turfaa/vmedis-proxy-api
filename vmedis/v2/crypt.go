package vmedisv2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
)

type Crypt struct {
	Key string
}

func NewCrypt(key string) *Crypt {
	return &Crypt{Key: key}
}

// EncryptToUnicode encrypts any value using Crypt.Encrypt, then encodes each byte of the encrypted output
// as its corresponding "\uXXXX" Unicode hexadecimal escape sequence.
func (c *Crypt) EncryptToUnicode(s any) (string, error) {
	encrypted, err := c.Encrypt(s)
	if err != nil {
		return "", fmt.Errorf("encrypt string: %w", err)
	}

	escaped := ""
	for _, b := range encrypted {
		escaped += fmt.Sprintf("\\u%04X", b)
	}

	return escaped, nil
}

// DecryptFromUnicode takes a string containing encrypted bytes in "\uXXXX" (or "\xHH") unicode/hexadecimal escapes
// (optionally mixed with normal characters), decodes them into encrypted bytes, and then decrypts those bytes into vPtr.
func (c *Crypt) DecryptFromUnicode(escaped string, vPtr any) error {
	var bytes []byte

	for i := 0; i < len(escaped); {
		// Try to match \uXXXX format
		if i+5 < len(escaped) && escaped[i] == '\\' && (escaped[i+1] == 'u' || escaped[i+1] == 'U') {
			var v int
			_, err := fmt.Sscanf(escaped[i+2:i+6], "%04X", &v)
			if err != nil {
				return fmt.Errorf("parse \\uXXXX at position %d: %w", i, err)
			}
			bytes = append(bytes, byte(v&0xFF))
			i += 6
		} else if i+3 < len(escaped) && escaped[i] == '\\' && (escaped[i+1] == 'x' || escaped[i+1] == 'X') {
			// Try to match \xHH format
			var v int
			_, err := fmt.Sscanf(escaped[i+2:i+4], "%02X", &v)
			if err != nil {
				return fmt.Errorf("parse \\xHH at position %d: %w", i, err)
			}
			bytes = append(bytes, byte(v))
			i += 4
		} else {
			// If not escaped, treat as a normal character byte.
			bytes = append(bytes, escaped[i])
			i++
		}
	}

	err := c.Decrypt(bytes, vPtr)
	if err != nil {
		return fmt.Errorf("decrypt bytes to %T: %w", vPtr, err)
	}

	return nil
}

// EncryptToURLEncoded encrypts any value using Crypt.Encrypt, then URL-escapes the encrypted bytes
// so the string can be safely transported in URLs.
func (c *Crypt) EncryptToURLEncoded(v any) (string, error) {
	encrypted, err := c.Encrypt(v)
	if err != nil {
		return "", fmt.Errorf("encrypt to URL encoded: %w", err)
	}

	return url.QueryEscape(string(encrypted)), nil
}

// DecryptFromURLEncodedToString takes a URL-encoded encrypted string, decodes and decrypts it, and returns the plaintext as a string.
func (c *Crypt) DecryptFromURLEncodedToString(urlEncoded string) (string, error) {
	var s string
	if err := c.DecryptFromURLEncoded(urlEncoded, &s); err != nil {
		return "", fmt.Errorf("decrypt from URL encoded to %T: %w", s, err)
	}

	return s, nil
}

// DecryptFromURLEncoded takes a URL-encoded encrypted string, decodes and decrypts it into vPtr.
func (c *Crypt) DecryptFromURLEncoded(urlEncoded string, vPtr any) error {
	encryptedStr, err := url.QueryUnescape(urlEncoded)
	if err != nil {
		return fmt.Errorf("unescape URL encoded: %w", err)
	}

	if err := c.Decrypt([]byte(encryptedStr), vPtr); err != nil {
		return fmt.Errorf("decrypt from URL encoded to %T: %w", vPtr, err)
	}

	return nil
}

// DecryptString decrypts the input encrypted byte slice and
// unmarshals the result into a string.
func (c *Crypt) DecryptString(data []byte) (string, error) {
	var s string
	if err := c.Decrypt(data, &s); err != nil {
		return "", fmt.Errorf("decrypt to string: %w", err)
	}

	return s, nil
}

// Encrypt takes an arbitrary object, marshals it to JSON, and encrypts it using a repeating-key XOR.
func (c *Crypt) Encrypt(v any) ([]byte, error) {
	plainBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal to JSON: %w", err)
	}

	plainStr := string(plainBytes)
	key := c.repeatKey(len(plainStr))
	out := make([]byte, len(plainStr))

	for i := 0; i < len(plainStr); i++ {
		out[i] = plainStr[i] ^ key[i]
	}

	return out, nil
}

// Decrypt takes an "encrypted" byte slice and unmarshals the decrypted JSON into vPtr.
// vPtr should be a pointer to the expected Go type.
func (c *Crypt) Decrypt(data []byte, vPtr any) error {
	if len(data) == 0 {
		if vPtr != nil {
			// set vPtr to its zero value using reflection
			rv := reflect.ValueOf(vPtr)
			if rv.Kind() == reflect.Ptr && !rv.IsNil() && rv.Elem().CanSet() {
				rv.Elem().Set(reflect.Zero(rv.Elem().Type()))
			}
		}
		return nil
	}

	cipherStr := string(data)
	key := c.repeatKey(len(cipherStr))
	decrypted := make([]byte, len(cipherStr))

	for i := 0; i < len(cipherStr); i++ {
		decrypted[i] = cipherStr[i] ^ key[i]
	}

	if err := json.Unmarshal(decrypted, vPtr); err != nil {
		log.Printf("failed to unmarshal to %T: %s", vPtr, string(decrypted))
		return fmt.Errorf("unmarshal to %T: %w", vPtr, err)
	}

	return nil
}

// repeatKey repeats the key to match the desired length, similar to JS .repeat() and .slice().
func (c *Crypt) repeatKey(length int) []byte {
	repeatCount := length / len(c.Key)
	if length%len(c.Key) != 0 {
		repeatCount++
	}
	repeated := ""
	for i := 0; i < repeatCount; i++ {
		repeated += c.Key
	}
	return []byte(repeated[:length])
}
