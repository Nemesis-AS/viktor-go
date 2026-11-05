package utils

import (
	"encoding/binary"
	"math/rand/v2"
	"net/netip"
	"strings"

	"github.com/Nemesis-AS/viktor-go/config"
)

func GenerateTransactionId() uint32 {
	return rand.Uint32()
}

func GeneratePeerId() SHA1Hash {
	var peerId SHA1Hash = [20]byte{}
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

func IntToIP(nn uint32) string {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], nn)
	return netip.AddrFrom4(bytes).String()
}
