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
	}

	announceUrl := data.Announce
	res, err := tracker.Connect(announceUrl)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)
}
