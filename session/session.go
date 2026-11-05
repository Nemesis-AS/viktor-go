package session

import (
	"fmt"
	"strconv"

	"github.com/Nemesis-AS/viktor-go/parser"
	"github.com/Nemesis-AS/viktor-go/peer"
	"github.com/Nemesis-AS/viktor-go/tracker"
	"github.com/Nemesis-AS/viktor-go/utils"
)

type PeerAddr struct {
	IpAddress uint32
	port      uint16
}

type TorrentSession struct {
	TorrentData parser.Torrent
	InfoHash    utils.SHA1Hash
	PeerId      utils.PeerIdType
	Key         uint32
	Downloaded  uint64
	Left        uint64
	Uploaded    uint64
	Trackers    []tracker.TrackerClient
	Peers       []peer.PeerConn
}

func (session *TorrentSession) CreateTorrentSession(torrent parser.Torrent, infoHash utils.SHA1Hash) error {
	session.TorrentData = torrent
	session.InfoHash = infoHash
	session.PeerId = utils.GeneratePeerId()
	session.Key = utils.GenerateKey()

	// @todo! Initialize the state as well

	return nil
}

func (session *TorrentSession) ConnectTrakers() error {
	var urls [][]string

	urls = append(urls, []string{session.TorrentData.Announce})
	urls = append(urls, session.TorrentData.AnnounceList...)

	for priority := range urls {
		for _, url := range urls[priority] {
			client, err := tracker.CreateTrackerClient(url)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = client.Connect()
			if err != nil {
				fmt.Println(err)
				continue
			}

			session.Trackers = append(session.Trackers, client)
		}
	}

	return nil
}

func (session *TorrentSession) GetPeers() ([]PeerAddr, error) {
	var peerMap map[uint64]byte = make(map[uint64]byte)
	var peers []PeerAddr

	for _, trackerConn := range session.Trackers {
		response, err := trackerConn.Announce(session.InfoHash, session.PeerId, session.Downloaded, session.Left, session.Uploaded, session.Key, tracker.EVENT_STARTED) // @todo! Make event dynamic
		if err != nil {
			fmt.Println("Announce returned an error: %w", err)
			continue
		}

		for idx, ip := range response.IpAddresses {
			hash := (uint64(ip) << 16) + uint64(response.TcpPorts[idx])
			if peerMap[hash] == 1 {
				continue
			}

			peerMap[hash] = 1
			peers = append(peers, PeerAddr{
				ip, response.TcpPorts[idx],
			})
		}
	}

	return peers, nil
}

func (session *TorrentSession) ConnectPeers(peers []PeerAddr) {
	for _, peerAddr := range peers {
		conn := peer.PeerConn{}
		ip := utils.IntToIP(peerAddr.IpAddress)
		addr := ip + ":" + strconv.Itoa(int(peerAddr.port))
		err := conn.Connect(addr, session.InfoHash, session.PeerId)
		if err != nil {
			// @todo! Add WARN log
			if conn.Connection != nil {
				conn.Connection.Close()
			}
			continue
		}

		session.Peers = append(session.Peers, conn)
	}
}

func (session *TorrentSession) Start() error {
	// Find Peers, Connect to them, and download the pieces. Also, service incoming requests

	// Step0: Check Disk for Torrent
	// @todo!

	// Step1: Connect to Trackers
	err := session.ConnectTrakers()
	if err != nil {
		return err
	}

	// Step2: Find the Peers
	peers, err := session.GetPeers()
	if err != nil {
		return err
	}

	// Step3: Connect to Peers
	session.ConnectPeers(peers)

	// Step4: Download Packets

	return nil
}
