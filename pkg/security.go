package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrMissing = errors.New("missing authorization header")
	ErrInvalid = errors.New("invalid API key")
)

// Principal représente l’identité extraite du token.
type Principal struct {
	KeyID string // identifiant safe (préfixe du hash)
}

// SecManager stocke les API keys autorisées (hashées).
type SecManager struct {
	allowed map[string]struct{} // hash hex -> vide
}

// NewManagerFromSlice construit un manager depuis des clés en clair.
func NewManagerFromSlice(keys []string) *SecManager {
	m := &SecManager{allowed: make(map[string]struct{})}
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		h := sha256.Sum256([]byte(k))
		hexHash := hex.EncodeToString(h[:])
		m.allowed[hexHash] = struct{}{}
	}
	return m
}

// ValidateBearer valide un header "Authorization: Bearer <token>"
func (m *SecManager) ValidateBearer(header string) (Principal, error) {
	if header == "" {
		return Principal{}, ErrMissing
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return Principal{}, ErrInvalid
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return Principal{}, ErrInvalid
	}

	// hash du token fourni
	h := sha256.Sum256([]byte(token))
	hexHash := hex.EncodeToString(h[:])

	if _, ok := m.allowed[hexHash]; !ok {
		return Principal{}, ErrInvalid
	}

	// KeyID = 8 premiers caractères du hash
	return Principal{KeyID: hexHash[:8]}, nil
}
