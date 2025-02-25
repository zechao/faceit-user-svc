package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/user"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("should hash password successfully", func(t *testing.T) {
		password := "superpassword"
		hashedPassword, err := user.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		assert.NoError(t, err)
	})

	t.Run("should return error when password too long", func(t *testing.T) {
		var password string
		for i := 0; i < 20; i++ {
			password += "password1234567890" // make the password longer than 72 characters
		}
		hashedPassword, err := user.HashPassword(password)
		assert.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
		assert.Empty(t, hashedPassword)
	})
}

func TestComparePassword(t *testing.T) {
	t.Run("should compare password successfully", func(t *testing.T) {
		password := "superpassword"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)

		assert.True(t, user.ComparePassword(string(hashedPassword), password))
	})
	t.Run("should return false when password does not match", func(t *testing.T) {
		password := "superpassword"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)

		assert.False(t, user.ComparePassword(string(hashedPassword), "wrongpassword"))
	})
	t.Run("should return false when hash is invalid", func(t *testing.T) {
		assert.False(t, user.ComparePassword(string("invalidhash"), "superpassword"))
	})
}
