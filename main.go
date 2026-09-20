package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path"

	bencode "github.com/jackpal/bencode-go"

	parser "github.com/Nemesis-AS/viktor-go/parser"
)

func main() {
	wd, _ := os.Getwd()
	path := path.Join(wd, "data", "big-buck-bunny.torrent")

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	var data parser.Torrent
	err = bencode.Unmarshal(file, &data)

	if err != nil {
		fmt.Println(err)
	}

	filecontent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	rawInfo, err := parser.ExtractInfo(filecontent)
	if err != nil {
		panic(err)
	}

	hash := sha1.Sum(rawInfo)

	fmt.Printf("Info hash: %x\n", hash)
	fmt.Println(data)
}
