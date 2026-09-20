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

const PROTOCOL_ID uint64 = 0x41727101980

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

type ConnectionResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func Unmarshal(data []byte) (ConnectionResponse, error) {
	if len(data) != 16 {
		return ConnectionResponse{}, errors.New("invalid response length")
	}

	return ConnectionResponse{
		Action:        binary.BigEndian.Uint32(data[0:4]),
		TransactionId: binary.BigEndian.Uint32(data[4:8]),
		ConnectionId:  binary.BigEndian.Uint64(data[8:16]),
	}, nil
}

type AnnounceRequest struct {
	ConnectionId  uint64
	Action        uint32
	TransactionId uint32
	InfoHash      [20]byte
	PeerId        [20]byte
	Downloaded    uint64
	Let           uint64
	Uploaded      uint64
	Event         uint32
	IpAddress     uint32
	Key           uint32
	NumWant       uint32
	Port          uint32
}

// @todo! Implement Scraping multiple torrents at once
type ScrapeRequest struct {
	ConnectionId  uint64
	Action        uint32
	TransactionId uint32
	InfoHash      [20]byte
}

func (request ScrapeRequest) Marshal() []byte {
	data := make([]byte, 36)

	binary.BigEndian.PutUint64(data[0:8], request.ConnectionId)
	binary.BigEndian.PutUint32(data[8:12], request.Action)
	binary.BigEndian.PutUint32(data[12:16], request.TransactionId)
	copy(data[16:36], request.InfoHash[:])

	return data
}

type ScrapeResponse struct {
	Action        uint32
	TransactionId uint32
	Seeders       uint32
	Completed     uint32
	Leechers      uint32
}

func (response *ScrapeResponse) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return errors.New("invalid response length")
	}

	fmt.Println("Buffer Length:", len(data))

	response.Action = binary.BigEndian.Uint32(data[0:4])
	response.TransactionId = binary.BigEndian.Uint32(data[4:8])
	response.Seeders = binary.BigEndian.Uint32(data[8:12])
	response.Completed = binary.BigEndian.Uint32(data[12:16])
	response.Leechers = binary.BigEndian.Uint32(data[16:20])

	return nil
}

type TrackerClient struct {
	Url          string
	Address      net.UDPAddr
	ConnectionId uint64
	InfoHash     [20]byte
}

func CreateTrackerClient(url string) (TrackerClient, error) {
	parts := strings.Split(url, "://")
	if len(parts) > 1 {
		if parts[0] != "udp" {
			return TrackerClient{}, errors.New("unknwon protocol found")
		}

		url = parts[1]
	}

	return TrackerClient{
		Url: url,
	}, nil
}

func (client *TrackerClient) Connect() error {
	addr, err := net.ResolveUDPAddr("udp", client.Url)
	if err != nil {
		return err
	}
	client.Address = *addr
	fmt.Println("Address:", addr)

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	body := ConnectionRequest{
		PROTOCOL_ID,
		0,
		generateTransactionId(),
	}
	fmt.Println("Request Bytes:", body.Marshal())

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	var buf []byte = make([]byte, 16)
	_, _, err = conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	fmt.Println("Response bytes:", buf)

	res, err := Unmarshal(buf)
	if err != nil {
		return err
	}

	fmt.Println("Response data:", res)
	client.ConnectionId = res.ConnectionId

	return nil
}

func (client *TrackerClient) Scrape(infoHash [20]byte) error {
	if client.ConnectionId == 0 {
		return errors.New("connection not established, please call Connect first")
	}

	conn, err := net.DialUDP("udp", nil, &client.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	body := ScrapeRequest{
		client.ConnectionId,
		2,
		generateTransactionId(),
		infoHash,
	}
	// fmt.Println("Scrape Request:", body)
	// fmt.Println("Scrape Bytes:", body.Marshal())

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	var buf = make([]byte, 500)
	size, _, err := conn.ReadFrom(buf)
	// fmt.Println("Scrape response size:", size)
	// fmt.Println("Scrape bytes received:", buf)

	var res ScrapeResponse = ScrapeResponse{}
	err = res.Unmarshal(buf[:size])
	if err != nil {
		return err
	}
	// fmt.Println("Scrape Response:", res)

	return nil
}

func generateTransactionId() uint32 {
	return uint32(rand.Int32())
}
