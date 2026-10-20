package main

import (
	"fmt"

	"github.com/Nemesis-AS/viktor-go/cli"
	"github.com/Nemesis-AS/viktor-go/parser"
	"github.com/Nemesis-AS/viktor-go/session"
)

func main() {
	torrentPath, err := cli.ParseArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := parser.Parse(torrentPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	state := session.TorrentSession{}
	err = state.CreateTorrentSession(data)
	// announceUrl := data.Announce
	// client, err := tracker.CreateTrackerClient(announceUrl)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.Connect()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	infoHash, err := parser.GetInfoHash(torrentPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	torrentState := parser.InitializeTorrentState(infoHash, data)

	// fmt.Println(len(state.TorrentData.Info.Pieces) / 20)
	fmt.Println(torrentState.TorrentRef.Info.Pieces)

	// err = client.Announce(torrentState, tracker.EVENT_NONE)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// err = client.Scrape(infoHash)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// peer := peer.PeerConn{}
	// err = peer.Connect("192.168.1.7:61326", torrentState)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
}
