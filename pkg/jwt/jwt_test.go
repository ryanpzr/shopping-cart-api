package jwtutil_test

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
)

func TestGenerate_Parse_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	token, err := jwtutil.Generate(1, "client", "test@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwtutil.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "client", claims.Role)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestParse_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	_, err := jwtutil.Parse("invalid.token.string")
	assert.Error(t, err)
}

func TestParse_WrongSecret(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-a")
	token, err := jwtutil.Generate(1, "client", "test@example.com")
	require.NoError(t, err)

	os.Setenv("JWT_SECRET", "secret-b")
	_, err = jwtutil.Parse(token)
	assert.Error(t, err)
}

func TestGenerate_MissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	token, err := jwtutil.Generate(1, "admin", "admin@example.com")
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestParse_MissingSecret_RejectsTokenSignedWithEmptyKey(t *testing.T) {
	// A token signed with an empty HMAC key must never be accepted.
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
		UserID:           1,
		Role:             "admin",
	})
	signed, err := forged.SignedString([]byte(""))
	require.NoError(t, err)

	t.Setenv("JWT_SECRET", "")
	_, err = jwtutil.Parse(signed)
	assert.Error(t, err)
}

func TestValidateSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	assert.ErrorIs(t, jwtutil.ValidateSecret(), jwtutil.ErrMissingSecret)

	t.Setenv("JWT_SECRET", "test-secret")
	assert.NoError(t, jwtutil.ValidateSecret())
}
