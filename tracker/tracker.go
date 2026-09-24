package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Nemesis-AS/viktor-go/config"
	"github.com/Nemesis-AS/viktor-go/parser"
	"github.com/Nemesis-AS/viktor-go/utils"
)

const PROTOCOL_ID uint64 = 0x41727101980

const (
	ACTION_CONNECT  uint32 = 0
	ACTION_ANNOUNCE uint32 = 1
	ACTION_SCRAPE   uint32 = 2
	ACTION_ERROR    uint32 = 3
)

const (
	EVENT_NONE      uint32 = 0
	EVENT_COMPLETED uint32 = 1
	EVENT_STARTED   uint32 = 2
	EVENT_STOPPED   uint32 = 3
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

type ConnectionResponse struct {
	Action        uint32
	TransactionId uint32
	ConnectionId  uint64
}

func (response *ConnectionResponse) Unmarshal(data []byte) error {
	if len(data) != 16 {
		return errors.New("invalid response length")
	}

	if binary.BigEndian.Uint32(data[0:4]) == 3 {
		return errors.New(string(data[4:]))
	}

	response.Action = binary.BigEndian.Uint32(data[0:4])
	response.TransactionId = binary.BigEndian.Uint32(data[4:8])
	response.ConnectionId = binary.BigEndian.Uint64(data[8:16])

	return nil
}

type AnnounceRequest struct {
	ConnectionId  uint64
	Action        uint32
	TransactionId uint32
	InfoHash      [20]byte
	PeerId        [20]byte
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	IpAddress     uint32
	Key           uint32
	NumWant       int32
	Port          uint16
}

func (request AnnounceRequest) Marshal() []byte {
	data := make([]byte, 98)

	binary.BigEndian.PutUint64(data[0:8], request.ConnectionId)
	binary.BigEndian.PutUint32(data[8:12], request.Action)
	binary.BigEndian.PutUint32(data[12:16], request.TransactionId)
	copy(data[16:36], request.InfoHash[:])
	copy(data[36:56], request.PeerId[:])
	binary.BigEndian.PutUint64(data[56:64], request.Downloaded)
	binary.BigEndian.PutUint64(data[64:72], request.Left)
	binary.BigEndian.PutUint64(data[72:80], request.Uploaded)
	binary.BigEndian.PutUint32(data[80:84], request.Event)
	binary.BigEndian.PutUint32(data[84:88], request.IpAddress)
	binary.BigEndian.PutUint32(data[88:92], request.Key)
	binary.BigEndian.PutUint32(data[92:96], uint32(request.NumWant))
	binary.BigEndian.PutUint16(data[96:98], request.Port)

	return data
}

type AnnounceResponse struct {
	Action        uint32
	TransactionId uint32
	Interval      uint32
	Leechers      uint32
	Seeders       uint32
	IpAddresses   []uint32
	TcpPorts      []uint16
}

func (response *AnnounceResponse) Unmarshal(data []byte) error {
	if len(data) < 20 {
		return errors.New("invalid response size")
	}

	response.Action = binary.BigEndian.Uint32(data[0:4])
	if response.Action == 3 {
		return errors.New(string(data[4:]))
	}

	response.TransactionId = binary.BigEndian.Uint32(data[4:8])
	response.Interval = binary.BigEndian.Uint32(data[8:12])
	response.Leechers = binary.BigEndian.Uint32(data[12:16])
	response.Seeders = binary.BigEndian.Uint32(data[16:20])

	if (len(data)-20)%6 != 0 {
		return errors.New("invalid peer list size")
	}

	for offset := 20; offset < len(data); offset += 6 {
		response.IpAddresses = append(response.IpAddresses, binary.BigEndian.Uint32(data[offset:offset+4]))
		response.TcpPorts = append(response.TcpPorts, binary.BigEndian.Uint16(data[offset+4:offset+6]))
	}

	return nil
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

	response.Action = binary.BigEndian.Uint32(data[0:4])
	if response.Action == 3 {
		return errors.New(string(data[4:]))
	}

	response.TransactionId = binary.BigEndian.Uint32(data[4:8])
	response.Seeders = binary.BigEndian.Uint32(data[8:12])
	response.Completed = binary.BigEndian.Uint32(data[12:16])
	response.Leechers = binary.BigEndian.Uint32(data[16:20])

	return nil
}

type TrackerClient struct {
	Url string
	// @todo! Revisit if this should be a pointer type
	Address      net.UDPAddr
	ConnectionId uint64
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

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	body := ConnectionRequest{
		PROTOCOL_ID,
		ACTION_CONNECT,
		utils.GenerateTransactionId(),
	}
	// fmt.Println("Request Bytes:", body.Marshal())

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = conn.SetReadDeadline(time.Now().Add(config.TRACKER_TIMEOUT_DURATION))
	if err != nil {
		return err
	}

	var buf []byte = make([]byte, 500)
	size, _, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	// fmt.Println("Response bytes:", buf)

	var res ConnectionResponse
	err = res.Unmarshal(buf[:size])
	if err != nil {
		return err
	}

	fmt.Println("Connect Response:", res)
	client.ConnectionId = res.ConnectionId

	return nil
}

func (client *TrackerClient) Announce(state parser.TorrentState, event uint32) error {
	if client.ConnectionId == 0 {
		return errors.New("connection not established, please call Connect first")
	}

	conn, err := net.DialUDP("udp", nil, &client.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	body := AnnounceRequest{
		client.ConnectionId,
		ACTION_ANNOUNCE,
		utils.GenerateTransactionId(),
		state.InfoHash,
		state.PeerId,
		state.Downloaded,
		state.Left,
		state.Uploaded,
		event,
		0,
		state.Key,
		-1,
		40000, // @todo! Update this based on the TCP connection
	}

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = conn.SetReadDeadline(time.Now().Add(config.TRACKER_TIMEOUT_DURATION))
	if err != nil {
		return err
	}

	buf := make([]byte, 500)
	size, _, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	var res AnnounceResponse = AnnounceResponse{}
	err = res.Unmarshal(buf[:size])
	if err != nil {
		return err
	}
	fmt.Println("Announce response:", res)

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
		ACTION_SCRAPE,
		utils.GenerateTransactionId(),
		infoHash,
	}
	// fmt.Println("Scrape Request:", body)
	// fmt.Println("Scrape Bytes:", body.Marshal())

	_, err = conn.Write(body.Marshal())
	if err != nil {
		return err
	}

	err = conn.SetReadDeadline(time.Now().Add(config.TRACKER_TIMEOUT_DURATION))
	if err != nil {
		return err
	}

	buf := make([]byte, 500)
	size, _, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}
	// fmt.Println("Scrape response size:", size)
	// fmt.Println("Scrape bytes received:", buf)

	var res ScrapeResponse = ScrapeResponse{}
	err = res.Unmarshal(buf[:size])
	if err != nil {
		return err
	}
	fmt.Println("Scrape Response:", res)

	return nil
}
