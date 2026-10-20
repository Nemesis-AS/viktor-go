package peer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/Nemesis-AS/viktor-go/config"
	"github.com/Nemesis-AS/viktor-go/parser"
)

const (
	PEER_MESSAGE_CHOKE          = 0
	PEER_MESSAGE_UNCHOKE        = 1
	PEER_MESSAGE_INTERESTED     = 2
	PEER_MESSAGE_NOT_INTERESTED = 3
	PEER_MESSAGE_HAVE           = 4
	PEER_MESSAGE_BITFIELD       = 5
	PEER_MESSAGE_REQUEST        = 6
	PEER_MESSAGE_PIECE          = 7
	PEER_MESSAGE_CANCEL         = 8
)

const BITTORRENT_SIGNATURE_LENGTH byte = 19

var BITTORRENT_SIGNATURE = [19]byte{
	'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't',
	' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l',
}
var RESERVED_BYTES = [8]byte{0}

type HandshakePacket struct {
	SignatureLength byte
	Signature       [19]byte
	Reserved        [8]byte
	InfoHash        [20]byte
	PeerId          [20]byte
}

func CreateHandshakeReq(infoHash [20]byte, peerId [20]byte) HandshakePacket {
	return HandshakePacket{
		SignatureLength: BITTORRENT_SIGNATURE_LENGTH,
		Signature:       BITTORRENT_SIGNATURE,
		Reserved:        RESERVED_BYTES,
		InfoHash:        infoHash,
		PeerId:          peerId,
	}
}

func (packet *HandshakePacket) Marshal() []byte {
	data := make([]byte, 68)

	data[0] = packet.SignatureLength
	copy(data[1:20], packet.Signature[:])
	copy(data[20:28], packet.Reserved[:])
	copy(data[28:48], packet.InfoHash[:])
	copy(data[48:68], packet.PeerId[:])

	return data
}

func (packet *HandshakePacket) Unmarshal(data []byte) (HandshakePacket, error) {
	if len(data) != 68 {
		return HandshakePacket{}, errors.New("invalid handshake packet")
	}

	return HandshakePacket{
		SignatureLength: data[0],
		Signature:       [19]byte(data[1:20]),
		Reserved:        [8]byte(data[20:28]),
		InfoHash:        [20]byte(data[28:48]),
		PeerId:          [20]byte(data[48:68]),
	}, nil
}

type PeerMessagePacket struct {
	Length  uint32
	Type    byte
	Payload []byte
}

func (pkt *PeerMessagePacket) Marshal() []byte {
	if pkt.Length == 0 {
		return make([]byte, 4)
	}

	var data []byte = make([]byte, pkt.Length+4)
	binary.BigEndian.PutUint32(data[0:4], pkt.Length)
	data[4] = pkt.Type
	copy(data[5:], pkt.Payload)

	return data
}

func (pkt *PeerMessagePacket) Unmarshal(data []byte) error {
	if len(data) < 5 {
		return errors.New("invalid message packet length")
	}

	pkt.Length = binary.BigEndian.Uint32(data[0:4])
	pkt.Type = data[4]
	pkt.Payload = data[4:]

	return nil
}

type PeerConn struct {
	Connection     *net.TCPConn
	PeerId         [20]byte // ID of the connected peer, not self
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
}

func (peer *PeerConn) Connect(peerUrl string, torrent parser.TorrentState) error {
	addr, err := net.ResolveTCPAddr("tcp", peerUrl)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	peer.Connection = conn

	err = peer.handshake(torrent.InfoHash, torrent.PeerId)
	if err != nil {
		conn.Close()
		peer.Connection = nil
		return err
	}

	// Connections start as Uninterested & Choked
	peer.AmChoking = true
	peer.AmInterested = false
	peer.PeerChoking = true
	peer.PeerInterested = false

	return nil
}

func (peer *PeerConn) handshake(infoHash [20]byte, peerId [20]byte) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	body := CreateHandshakeReq(infoHash, peerId)

	_, err := peer.Connection.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = peer.Connection.SetReadDeadline(time.Now().Add(config.PEER_HANDSHAKE_TIMEOUT_DURATION))
	if err != nil {
		return err
	}

	var buf = make([]byte, 68)
	_, err = io.ReadFull(peer.Connection, buf)
	if err != nil {
		return err
	}

	var res HandshakePacket
	res, err = res.Unmarshal(buf[:])
	if err != nil {
		return err
	}

	if res.InfoHash != body.InfoHash {
		return errors.New("invalid handshake: info hashes do not match")
	}

	peer.PeerId = res.PeerId

	fmt.Printf("Handshake Successful with Peer ID: %v\n", peer.PeerId)

	return nil
}

func (peer *PeerConn) Choke() error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	pkt := PeerMessagePacket{
		Length: 1,
		Type:   PEER_MESSAGE_CHOKE,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	peer.AmChoking = true

	return nil
}

func (peer *PeerConn) Unchoke() error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	pkt := PeerMessagePacket{
		Length: 1,
		Type:   PEER_MESSAGE_UNCHOKE,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	peer.AmChoking = false

	return nil
}

func (peer *PeerConn) Interested() error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	pkt := PeerMessagePacket{
		Length: 1,
		Type:   PEER_MESSAGE_INTERESTED,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	peer.AmInterested = true

	return nil
}

func (peer *PeerConn) NotInterested() error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	pkt := PeerMessagePacket{
		Length: 1,
		Type:   PEER_MESSAGE_NOT_INTERESTED,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	peer.AmInterested = false

	return nil
}

func (peer *PeerConn) MarkHave(pieceIdx uint32) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, pieceIdx)

	pkt := PeerMessagePacket{
		Length:  5, // 4 byte number + 1 byte type
		Type:    PEER_MESSAGE_HAVE,
		Payload: payload,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (peer *PeerConn) SendBitfield(state parser.TorrentState) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	pieceCount := uint32(len(state.TorrentRef.Info.Pieces) / 20)
	bitfield := make([]byte, uint32(math.Ceil(float64(pieceCount)/8)))

	// @todo! Read disk for correct data

	pkt := PeerMessagePacket{
		Length:  uint32(1 + len(bitfield)),
		Type:    PEER_MESSAGE_BITFIELD,
		Payload: bitfield,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (peer *PeerConn) RequestPiece(pieceIdx uint32, begin uint32, length uint32) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], pieceIdx)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)

	pkt := PeerMessagePacket{
		Length:  13, // 1 byte type + 3 x 4 byte integers
		Type:    PEER_MESSAGE_REQUEST,
		Payload: payload,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (peer *PeerConn) SendPiece(pieceIdx uint32, begin uint32, data []byte) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	payload := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(payload[0:4], pieceIdx)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	copy(payload[8:], data)

	pkt := PeerMessagePacket{
		Length:  uint32(1 + len(payload)),
		Type:    PEER_MESSAGE_PIECE,
		Payload: payload,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (peer *PeerConn) CancelPieceReq(pieceIdx uint32, begin uint32, length uint32) error {
	if peer.Connection == nil {
		return errors.New("connection not initialized yet")
	}

	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], pieceIdx)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	binary.BigEndian.PutUint32(payload[8:12], length)

	pkt := PeerMessagePacket{
		Length:  13, // 1 byte type + 3 x 4 bit integers
		Type:    PEER_MESSAGE_CANCEL,
		Payload: payload,
	}

	_, err := peer.Connection.Write(pkt.Marshal())
	if err != nil {
		return err
	}

	return nil
}
