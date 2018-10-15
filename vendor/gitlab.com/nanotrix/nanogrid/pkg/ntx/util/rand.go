package util

import "math/rand"

const charsetFroRandString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charsetFroRandString[rand.Intn(len(charsetFroRandString))]
	}
	return string(b)
}
