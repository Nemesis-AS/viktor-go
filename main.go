package main

import (
	"fmt"

	"github.com/Nemesis-AS/viktor-go/cli"
	"github.com/Nemesis-AS/viktor-go/parser"
	"github.com/Nemesis-AS/viktor-go/tracker"
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

	announceUrl := data.Announce
	client, err := tracker.CreateTrackerClient(announceUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	infoHash, err := parser.GetInfoHash(torrentPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	torrentState := parser.InitializeTorrentState(infoHash, data)

	err = client.Announce(torrentState, tracker.EVENT_NONE)
	if err != nil {
		fmt.Println(err)
		return
	}

	// err = client.Scrape(infoHash)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
}
