package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strings"
	"time"
)

type ConnectionRequest struct {
	ProtocolId    uint64
	Action        uint32
	TransactionId uint32
}

func (req ConnectionRequest) Marshal() []byte {
	data := make([]byte, 16)

	binary.BigEndian.PutUint64(data[0:8], req.ProtocolId)
	binary.BigEndian.PutUint32(data[8:12], req.Action)
	binary.BigEndian.PutUint32(data[12:16], req.TransactionId)

	return data
}

type AnnounceRequest struct {
	ConnectionId  uint64
	Action        uint32
	TransactionId uint32
	InfoHash      string
	PeerId        string
	Downloaded    uint64
	Let           uint64
	Uploaded      uint64
	Event         uint32
	IpAddress     uint32
	Key           uint32
	NumWant       uint32
	Port          uint32
}

type ConnectionResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func Unmarshal(data []byte) (ConnectionResponse, error) {
	if len(data) < 16 {
		return ConnectionResponse{}, errors.New("data length less than struct")
	}

	return ConnectionResponse{
		Action:        uint32(binary.BigEndian.Uint32(data[0:4])),
		TransactionId: uint32(binary.BigEndian.Uint32(data[4:8])),
		ConnectionId:  uint64(binary.BigEndian.Uint64(data[8:16])),
	}, nil
}

func Connect(url string) (ConnectionResponse, error) {
	txnId := rand.Int32()

	body := ConnectionRequest{
		0x41727101980,
		0,
		uint32(txnId),
	}

	fmt.Println("Request Bytes:", body.Marshal())

	// return ConnectionResponse{}, nil

	parts := strings.Split(url, "://")
	if len(parts) > 1 {
		if parts[0] != "udp" {
			return ConnectionResponse{}, errors.New("unknwon protocol found")
		}

		url = parts[1]
	}

	fmt.Println(url)

	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		return ConnectionResponse{}, err
	}

	fmt.Println("Address:", addr)

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return ConnectionResponse{}, err
	}
	defer conn.Close()

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return ConnectionResponse{}, err
	}

	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return ConnectionResponse{}, err
	}

	var buf []byte = make([]byte, 16)
	_, _, err = conn.ReadFrom(buf)
	if err != nil {
		return ConnectionResponse{}, err
	}

	fmt.Println("Response bytes:", buf)

	res, err := Unmarshal(buf)
	if err != nil {
		return ConnectionResponse{}, err
	}

	fmt.Println("Response data:", res)

	return res, nil
}
