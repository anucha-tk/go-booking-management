package auth_test

import (
	"testing"

	"go-booking-management-init/pkg/auth"

	"github.com/stretchr/testify/assert"
)

func TestArgon2id_HashAndVerify(t *testing.T) {
	password := "SecurePass123!"

	hash, err := auth.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	match, err := auth.VerifyPassword(password, hash)
	assert.NoError(t, err)
	assert.True(t, match)
}
