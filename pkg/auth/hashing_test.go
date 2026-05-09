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

func TestHashPasswordWithConfig(t *testing.T) {
	t.Run("custom config", func(t *testing.T) {
		hash, err := auth.HashPasswordWithConfig("mypassword", &auth.Config{
			Memory:      64 * 1024,
			Iterations:  2,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("empty password", func(t *testing.T) {
		hash, err := auth.HashPasswordWithConfig("", auth.DefaultConfig)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func TestVerifyPassword_EdgeCases(t *testing.T) {
	t.Run("wrong password", func(t *testing.T) {
		hash, _ := auth.HashPassword("correct-password")
		match, err := auth.VerifyPassword("wrong-password", hash)
		assert.NoError(t, err)
		assert.False(t, match)
	})

	t.Run("invalid hash format", func(t *testing.T) {
		match, err := auth.VerifyPassword("password", "invalid-hash")
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidHash, err)
		assert.False(t, match)
	})

	t.Run("empty hash", func(t *testing.T) {
		match, err := auth.VerifyPassword("password", "")
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidHash, err)
		assert.False(t, match)
	})

	t.Run("malformed hash parts count", func(t *testing.T) {
		match, err := auth.VerifyPassword("password", "$argon2id$v=19$m=65536")
		assert.Error(t, err)
		assert.False(t, match)
	})

	t.Run("incompatible argon2 version", func(t *testing.T) {
		hash := "$argon2id$v=999$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0$dGVzdGhhc2g="
		match, err := auth.VerifyPassword("password", hash)
		assert.Error(t, err)
		assert.Equal(t, auth.ErrIncompatibleVersion, err)
		assert.False(t, match)
	})

	t.Run("invalid base64 salt", func(t *testing.T) {
		hash := "$argon2id$v=19$m=65536,t=3,p=4$!!!$dGVzdGhhc2g="
		match, err := auth.VerifyPassword("password", hash)
		assert.Error(t, err)
		assert.False(t, match)
	})

	t.Run("invalid base64 hash part", func(t *testing.T) {
		hash := "$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0$!!!"
		match, err := auth.VerifyPassword("password", hash)
		assert.Error(t, err)
		assert.False(t, match)
	})

	t.Run("invalid version format", func(t *testing.T) {
		hash := "$argon2id$v=abc$m=65536,t=3,p=4$salt$hash"
		match, err := auth.VerifyPassword("password", hash)
		assert.Error(t, err)
		assert.False(t, match)
	})

	t.Run("invalid config format", func(t *testing.T) {
		hash := "$argon2id$v=19$m=abc,t=3,p=4$salt$hash"
		match, err := auth.VerifyPassword("password", hash)
		assert.Error(t, err)
		assert.False(t, match)
	})
}

func TestComparePassword(t *testing.T) {
	password := "mypassword"
	hash, _ := auth.HashPassword(password)

	t.Run("match", func(t *testing.T) {
		err := auth.ComparePassword(hash, password)
		assert.NoError(t, err)
	})

	t.Run("no match", func(t *testing.T) {
		err := auth.ComparePassword(hash, "wrong")
		assert.Error(t, err)
		assert.Equal(t, "password does not match", err.Error())
	})

	t.Run("invalid hash", func(t *testing.T) {
		err := auth.ComparePassword("invalid", password)
		assert.Error(t, err)
	})
}
