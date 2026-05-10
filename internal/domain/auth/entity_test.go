package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRole_Constants(t *testing.T) {
	assert.Equal(t, UserRole("admin"), RoleAdmin)
	assert.Equal(t, UserRole("customer"), RoleCustomer)
	assert.Equal(t, UserRole("officer"), RoleOfficer)
	assert.Equal(t, UserRole("guest"), RoleGuest)
}

func TestUser_Struct(t *testing.T) {
	u := User{
		ID:           1,
		Email:        "test@test.com",
		PasswordHash: "secret",
		Role:         RoleAdmin,
	}

	assert.Equal(t, int32(1), u.ID)
	assert.Equal(t, "test@test.com", u.Email)
	assert.Equal(t, "secret", u.PasswordHash)
	assert.Equal(t, RoleAdmin, u.Role)
	assert.True(t, u.CreatedAt.IsZero())
	assert.True(t, u.UpdatedAt.IsZero())
}
