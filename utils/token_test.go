package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() AppConfig {
	return AppConfig{SecretKey: strings.Repeat("s", 32), EncryptionKey: strings.Repeat("e", 32)}
}

func TestTokenRoundTrip(t *testing.T) {
	cfg := testConfig()
	tok, err := IssueToken(cfg, Claims{ID: "u1", Email: "a@b.c", Role: "candidate", Type: TokenAccess}, time.Minute)
	require.NoError(t, err)

	claims, err := ParseToken(cfg, tok)
	require.NoError(t, err)
	assert.Equal(t, "u1", claims.ID)
	assert.Equal(t, "candidate", claims.Role)
	assert.Equal(t, TokenAccess, claims.Type)
}

func TestParseTokenRejects(t *testing.T) {
	cfg := testConfig()
	expired, _ := IssueToken(cfg, Claims{ID: "u1"}, -time.Minute)
	good, _ := IssueToken(cfg, Claims{ID: "u1"}, time.Minute)
	other := cfg
	other.SecretKey = strings.Repeat("x", 32)
	otherEnc := cfg
	otherEnc.EncryptionKey = strings.Repeat("y", 32)

	for name, tc := range map[string]struct {
		cfg   AppConfig
		token string
	}{
		"expired":          {cfg, expired},
		"garbage":          {cfg, "not-a-token"},
		"wrong encryption": {otherEnc, good},
		"empty":            {cfg, ""},
	} {
		_, err := ParseToken(tc.cfg, tc.token)
		assert.Error(t, err, name)
	}
	_ = other
}

func TestBearerToken(t *testing.T) {
	tok, ok := BearerToken("Bearer abc")
	assert.True(t, ok)
	assert.Equal(t, "abc", tok)
	_, ok = BearerToken("Basic abc")
	assert.False(t, ok)
	_, ok = BearerToken("")
	assert.False(t, ok)
}
