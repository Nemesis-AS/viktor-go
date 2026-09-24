package peer

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Nemesis-AS/viktor-go/config"
	"github.com/Nemesis-AS/viktor-go/parser"
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

type PeerConn struct {
	Connection *net.TCPConn
	PeerId     [20]byte // ID of the connected peer, not self
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
		return err
	}

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

	// @todo! May require io.FullRead()
	var buf = make([]byte, 500)
	n, err := peer.Connection.Read(buf)
	if err != nil {
		return err
	}

	var res HandshakePacket
	res, err = res.Unmarshal(buf[:n])
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
