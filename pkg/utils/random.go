package utils

import (
	"math/rand"
)

var (
	character = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	chLen     = len(character)
)

func RandomString(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = character[rand.Intn(chLen)]
	}
	return string(b)
}
