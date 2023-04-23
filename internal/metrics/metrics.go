package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func (m *Metrics) ValidMAC(key []byte) bool {
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

func (m *Metrics) GetSign(key []byte) ([]byte, error) {
	var (
		message []byte
		sign    []byte
	)

	switch m.MType {
	case "gauge":
		if m.Value == nil {
			return nil, fmt.Errorf("invalid value")
		}
		message = []byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value))

	case "counter":
		if m.Delta == nil {
			return nil, fmt.Errorf("invalid value")
		}
		message = []byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta))
	}

	sign, err := Sign(message, key)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

func (m *Metrics) Sign(key []byte) error {
	sign, err := m.GetSign(key)
	if err != nil {
		return err
	}

	m.Hash = hex.EncodeToString(sign)
	return nil
}

func Sign(message, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(message)
	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}
