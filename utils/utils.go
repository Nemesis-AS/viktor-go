package utils

import (
	"math/rand/v2"
	"strings"

	"github.com/Nemesis-AS/viktor-go/config"
)

func GenerateTransactionId() uint32 {
	return rand.Uint32()
}

func GeneratePeerId() [20]byte {
	var peerId [20]byte = [20]byte{}
	charset := "abcdeffghijklmnopqrstuvwxyz1234567890"

	peerId[0] = '-'
	copy(peerId[1:3], []byte(config.CLIENT_PREFIX))
	copy(peerId[3:6], []byte(strings.ReplaceAll(config.CLIENT_VERSION, ".", "")))
	peerId[6] = '0'
	peerId[7] = '-'

	for idx := 8; idx < 20; idx++ {
		ch := rand.IntN(len(charset))
		peerId[idx] = charset[ch]
	}

	return peerId
}

func GenerateKey() uint32 {
	return rand.Uint32()
}
