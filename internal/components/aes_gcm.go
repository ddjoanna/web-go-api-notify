package component

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	shared "notify-service/internal"

	log "github.com/sirupsen/logrus"
)

// validKeyLengths 定義允許的 AES 金鑰長度
var validKeyLengths = map[int]bool{
	16: true,
	24: true,
	32: true,
}

// AES-GCM 結構，封裝 AES-GCM 加密/解密功能
type AesGcm struct {
	gcm cipher.AEAD
}

// NewAesGcm 初始化 AES-GCM 模式
func NewAesGcm(
	config *shared.Config,
) (*AesGcm, error) {
	if !validKeyLengths[len(config.AESKey)] {
		return nil, fmt.Errorf("invalid AES key length: must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher([]byte(config.AESKey))
	if err != nil {
		log.WithError(err).Error("cipher initialization failed")
		return nil, fmt.Errorf("cipher initialization failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.WithError(err).Error("GCM mode initialization failed")
		return nil, fmt.Errorf("GCM mode initialization failed: %w", err)
	}

	return &AesGcm{gcm: gcm}, nil
}

// AES Encrypt 使用 GCM 模式加密訊息
func (a *AesGcm) AesEncrypt(plaintext string) (string, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.WithError(err).Error("nonce generation failed")
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	ciphertext := a.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// AES Decrypt 解密使用 GCM 模式加密的訊息
func (a *AesGcm) AesDecrypt(ciphertextHex string) (string, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		log.WithError(err).Error("hex decoding failed")
		return "", fmt.Errorf("hex decoding failed: %w", err)
	}

	if len(ciphertext) < a.gcm.NonceSize() {
		log.Error("invalid ciphertext: shorter than nonce length")
		return "", fmt.Errorf("invalid ciphertext: shorter than nonce length")
	}

	nonce, ciphertext := ciphertext[:a.gcm.NonceSize()], ciphertext[a.gcm.NonceSize():]
	plaintext, err := a.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.WithError(err).Error("decryption failed")
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
