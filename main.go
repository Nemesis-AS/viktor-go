package main

import (
	"fmt"

	"github.com/Nemesis-AS/viktor-go/cli"
	"github.com/Nemesis-AS/viktor-go/parser"
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

	fmt.Println(data)
}
