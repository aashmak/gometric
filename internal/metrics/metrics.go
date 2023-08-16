// Пакет metrics описывает структуру хранения метрик и методы.
package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Metrics описывает структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// ValidMAC проверяет подпись.
// Переподписывает с помощью ключа и сравнивает полученный Hash с существующим.
func (m *Metrics) ValidMAC(key string) bool {
	var (
		mac1 []byte
		mac2 []byte
	)

	mac1, err := m.GetSign(key)
	if err != nil {
		return false
	}

	mac2, err = hex.DecodeString(m.Hash)
	if err != nil {
		return false
	}

	return hmac.Equal(mac1, mac2)
}

// GetSign подписывает с помощью ключа, возвращая значение Hash
func (m *Metrics) GetSign(key string) ([]byte, error) {
	var (
		message string
		sign    []byte
	)

	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return nil, fmt.Errorf("invalid value")
		}
		message = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)

	case "counter":
		if m.Delta == nil {
			return nil, fmt.Errorf("invalid value")
		}
		message = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	}

	sign, err := Sign(message, key)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

// Sign подписывает с помощью ключа, сохраняя значение в Hash
func (m *Metrics) Sign(key string) error {
	sign, err := m.GetSign(key)
	if err != nil {
		return err
	}

	m.Hash = hex.EncodeToString(sign)
	return nil
}

// Sign подписывает с помощью ключа, возвращая Hash
func Sign(message, key string) ([]byte, error) {

	mac := hmac.New(sha256.New, []byte(key))
	_, err := mac.Write([]byte(message))
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}
