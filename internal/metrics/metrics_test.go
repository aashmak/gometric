package metrics

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHash1(t *testing.T) {
	message := "test message"
	key := "secret"

	hash, err := Sign(message, key)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if hex.EncodeToString(hash) != "3bcebf43c85d20bba6e3b6ba278af1d2ba3ab0d57de271b0ad30b833e851c5a6" {
		t.Errorf("Error: hash is not valid")
	}
}

func TestHash2(t *testing.T) {
	message := ""
	key := ""
	_, err := Sign(message, key)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if message == "" {
		fmt.Printf("msg: %s\n", message)
	}
}

func TestGetSign(t *testing.T) {
	key := "secret"

	tests := []struct {
		name string
		msg  string
		hash string
	}{
		{
			name: "test sign gauge type #1",
			msg:  `{"id":"Alloc","type":"gauge","value":1907608}`,
			hash: "eedda5f934b51f71d9255066a51e63a30567c14960c9c5922a8389adac649819",
		},
		{
			name: "test sign gauge type #2",
			msg:  `{"id":"BuckHashSys","type":"gauge","value":3877}`,
			hash: "8050e62e1c10435ccccd85ad83aba55336a28a950413bc3fec359f1d18c2b45a",
		},
		{
			name: "test sign gauge type #3",
			msg:  `{"id":"Frees","type":"gauge","value":2925}`,
			hash: "b1986b52e14e9713339b0d5abac15c72f4fdabff7fd737333ccee84f3f344ed2",
		},
		{
			name: "test sign counter type #4",
			msg:  `{"id":"PollCount","type":"counter","delta":1}`,
			hash: "ce97c6062da4477a5fad4cfdd24f0f24e474d309b1f054928dd138683d1cab12",
		},
		{
			name: "test sign counter type #4",
			msg:  `{"id":"PollCount","type":"counter","delta":100}`,
			hash: "f88b6a2173d62673e9d7155b63882abdae388534d20ff569d24dd1df1132bb8b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metrics{}

			err := json.Unmarshal([]byte(tt.msg), &m)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			sign, err := m.GetSign(key)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			if hex.EncodeToString(sign) != tt.hash {
				t.Errorf("Error: hash is not valid")
			}
		})
	}
}

func TestValidMAC(t *testing.T) {
	key := "secret"

	tests := []struct {
		name     string
		msg      string
		validMAC bool
	}{
		{
			name: "test validMAC #1",
			msg:  `{"id":"Alloc","type":"gauge","value":1907608,"hash":"eedda5f934b51f71d9255066a51e63a30567c14960c9c5922a8389adac649819"}`,
		},
		{
			name: "test validMAC #2",
			msg:  `{"id":"BuckHashSys","type":"gauge","value":3877,"hash":"8050e62e1c10435ccccd85ad83aba55336a28a950413bc3fec359f1d18c2b45a"}`,
		},
		{
			name: "test validMAC #2",
			msg:  `{"id":"GCCPUFraction","type":"gauge","value":0.02452708223955301,"hash":"0e2a4947c8de1880d7dd233ac1a874f296c73bec044b49f318bdd19e876fddd1"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metrics{}

			err := json.Unmarshal([]byte(tt.msg), &m)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			if !m.ValidMAC(key) {
				t.Errorf("Error: hash not valid")
			}
		})
	}
}

func TestSign(t *testing.T) {
	key := "secret"

	tests := []struct {
		name string
		msg  string
		hash string
	}{
		{
			name: "test Sign #1",
			msg:  `{"id":"Alloc","type":"gauge","value":1907608}`,
			hash: "eedda5f934b51f71d9255066a51e63a30567c14960c9c5922a8389adac649819",
		},
		{
			name: "test Sign #2",
			msg:  `{"id":"BuckHashSys","type":"gauge","value":3877}`,
			hash: "8050e62e1c10435ccccd85ad83aba55336a28a950413bc3fec359f1d18c2b45a",
		},
		{
			name: "test Sign #3",
			msg:  `{"id":"Frees","type":"gauge","value":2925}`,
			hash: "b1986b52e14e9713339b0d5abac15c72f4fdabff7fd737333ccee84f3f344ed2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metrics{}

			err := json.Unmarshal([]byte(tt.msg), &m)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			if err := m.Sign(key); err != nil {
				t.Errorf("Error sign")
			}

			if m.Hash != tt.hash {
				t.Errorf("Error: incorrect hash")
			}
		})
	}
}

func ExampleSign() {

	message := "test message"
	key := "secret"

	hash, _ := Sign(message, key)
	fmt.Printf("%s", hex.EncodeToString(hash))

	// Output:
	// 3bcebf43c85d20bba6e3b6ba278af1d2ba3ab0d57de271b0ad30b833e851c5a6
}

func ExampleMetrics_ValidMAC() {
	m := Metrics{}
	message := `{"id":"Alloc","type":"gauge","value":1907608,"hash":"eedda5f934b51f71d9255066a51e63a30567c14960c9c5922a8389adac649819"}`
	key := "secret"

	json.Unmarshal([]byte(message), &m)

	fmt.Printf("%t", m.ValidMAC(key))

	// Output:
	// true
}
