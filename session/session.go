package session

import (
	"fmt"

	"github.com/Nemesis-AS/viktor-go/parser"
	"github.com/Nemesis-AS/viktor-go/peer"
	"github.com/Nemesis-AS/viktor-go/tracker"
)

type TorrentSession struct {
	TorrentData parser.Torrent
	Peer        []peer.PeerConn
	Trackers    []tracker.TrackerClient
}

func (session *TorrentSession) CreateTorrentSession(torrent parser.Torrent) error {
	session.TorrentData = torrent
	return nil
}

func (session *TorrentSession) ConnectTrakers() error {
	var urls [][]string

	urls = append(urls, []string{session.TorrentData.Announce})
	urls = append(urls, session.TorrentData.AnnounceList...)

	for priority := range urls {
		fmt.Println(priority)
	}

	return nil
}
