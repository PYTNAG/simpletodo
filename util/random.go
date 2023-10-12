package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz_()!$"

type FullUserInfo struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Hash     []byte `json:"hash"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomUsername generates a random string of length 24
func RandomUsername() string {
	return RandomString(24)
}

// RandomPassword generates a random string of length 16
func RandomPassword() string {
	return RandomString(16)
}

// RandomID generates a random positive int32
func RandomID() int32 {
	res := rand.Int31()
	if res == int32(0) {
		res = 1
	}

	return res
}

func RandomUser() (FullUserInfo, error) {
	pass := RandomPassword()
	hash, err := HashPassword(pass)

	if err != nil {
		return FullUserInfo{}, err
	}

	return FullUserInfo{
		ID:       RandomID(),
		Username: RandomUsername(),
		Password: pass,
		Hash:     hash,
	}, nil
}
