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

	infoHash, err := parser.GetInfoHash(torrentPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	session := session.TorrentSession{}
	err = session.CreateTorrentSession(data, infoHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = session.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
}
