package peer

// import (
// 	"encoding/binary"
// 	"net"
// )

// const (
// 	BITTORRENT_SIGNATURE_LENGTH byte = 19
// 	BITTORRENT_SIGNATURE = "BitTorrent protocol"
// 	RESERVED_BYTES = [8]byte{}
// )

// type Peer struct {
// 	Connection net.Conn
// }

// func (peer *Peer) Connect(url string) error {
// 	conn, err := net.Dial("udp", url)
// 	if err != nil {
// 		return err
// 	}

// 	peer.Connection = conn

// 	return nil
// }

// type HandshakeRequest struct {
// 	SignatureLength byte
// 	Signature       [19]byte
// 	Reserved        [8]byte
// 	InfoHash        [20]byte
// 	PeerId          [20]byte
// }

// func CreateHandshakeReq() (HandshakeRequest, error) {
// 	return HandshakeRequest{
// 		SignatureLength: 19,
// 	}
// }

// func (req *HandshakeRequest) Marshal() ([]byte, error) {
// 	data := make([]byte, 71)

// 	binary.BigEndian.PutUint32(data[0:4], req.SignatureLength)
// 	copy(data[4:23], req.Signature[:])
// 	copy(data[23:31], req.Reserved[:])
// 	copy(data[31:51], req.InfoHash[:])
// 	copy(data[51:71], req.PeerId[:])

// 	return data, nil
// }

// func (peer *Peer) handshake() error {

// }
