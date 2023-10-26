package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := RandomPassword()

	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = CheckPassword(password, hash)
	require.NoError(t, err)

	wrongPassword := RandomPassword()
	err = CheckPassword(wrongPassword, hash)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	repeatedHash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, repeatedHash)
	require.NotEqual(t, hash, repeatedHash)
}

func TestTooLongPassword(t *testing.T) {
	password := RandomString(73)

	hash, err := HashPassword(password)
	require.Error(t, err)
	require.Empty(t, hash)
}
