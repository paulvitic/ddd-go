package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestConfiguration struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port"`
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestProperties(t *testing.T) {

	tests := []struct {
		name    string
		profile string
		wantErr bool
	}{
		{name: "Default profile empty", profile: "", wantErr: false},
		{name: "Valid profile suffix", profile: "dev", wantErr: false},
		{name: "Encrypted", profile: "sops", wantErr: false},
		{name: "Non-existent profile suffix", profile: "invalid", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
				Properties[TestConfiguration](tc.profile)
			} else {
				result := Properties[TestConfiguration](tc.profile)
				assert.Equal(t, 3000, result.Port, "Port should be 3000")
			}
		})
	}
}
