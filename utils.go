package main

import "math/rand/v2"

const idAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateID(minLen, maxLen int) string {
	if maxLen < minLen {
		minLen, maxLen = maxLen, minLen
	}
	length := minLen
	if maxLen > minLen {
		length = rand.IntN(maxLen-minLen+1) + minLen
	}

	id := make([]byte, length)
	for i := range id {
		id[i] = idAlphabet[rand.IntN(len(idAlphabet))]
	}
	return string(id)
}
