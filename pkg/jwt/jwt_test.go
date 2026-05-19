package jwtutil_test

import (
	"os"
	"testing"

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
