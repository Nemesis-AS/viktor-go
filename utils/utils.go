package utils

import "math/rand/v2"

func GenerateTransactionId() uint32 {
	return rand.Uint32()
}

// @todo! Implement proper peer id format according to BEP0020
func GeneratePeerId() [20]byte {
	var peerId [20]byte = [20]byte{}
	charset := "abcdeffghijklmnopqrstuvwxyz1234567890"

	for idx := range 20 {
		ch := rand.IntN(len(charset))
		peerId[idx] = charset[ch]
	}

	return peerId
}

func GenerateKey() uint32 {
	return rand.Uint32()
}
