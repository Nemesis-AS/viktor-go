package parser

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	bencode "github.com/jackpal/bencode-go"

	"github.com/Nemesis-AS/viktor-go/utils"
)

type Torrent struct {
	Announce     string
	AnnounceList [][]string `bencode:"announce-list"`
	CreationDate uint64     `becode:"creation date"`
	CreatedBy    string     `bencode:"created by"`
	Comment      string     `bencode:"comment"`
	Info         TorrentInfo
}

func Parse(filepath string) (Torrent, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return Torrent{}, err
	}
	defer file.Close()

	var data Torrent
	err = bencode.Unmarshal(file, &data)
	if err != nil {
		return Torrent{}, err
	}

	return data, nil
}

func (torrent Torrent) String() string {
	size := (float32(len(torrent.Info.Pieces)/20) * float32(torrent.Info.PieceLength)) / (1024 * 1024)

	var announceList []string
	for _, announce := range torrent.AnnounceList {
		announceList = append(announceList, strings.Join(announce, ", "))
	}
	formattedAnnounceList := strings.Join(announceList, "\n\t\t")

	return fmt.Sprintf("Announce:\t%v\nAnnounce List:\t%v\nCreation Date:\t%v\nCreated By:\t%v\nComment:\t%v\nName:\t\t%v\nPiece Length:\t%v\nPiece Count:\t%v\nTorrent Size:\t%0.2fMiB\n", torrent.Announce, formattedAnnounceList, torrent.CreationDate, torrent.CreatedBy, torrent.Comment, torrent.Info.Name, torrent.Info.PieceLength, len(torrent.Info.Pieces), size)
}

type TorrentInfo struct {
	Name        string
	PieceLength uint `bencode:"piece length"`
	Length      uint
	Files       []TorrentFiles
	Pieces      string
}

type TorrentFiles struct {
	Length uint
	Path   []string
}

func ExtractInfoRaw(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("could not open file")
	}

	marker := []byte("4:info")
	pos := bytes.Index(data, marker)
	if pos == -1 {
		return nil, errors.New("info dictionary not found")
	}

	start := pos + len(marker)

	if start >= len(data) || data[start] != 'd' {
		return nil, errors.New("info is not a dictionary")
	}

	depth := 0

	for i := start; i < len(data); {
		switch data[i] {
		case 'd', 'l':
			depth++
			i++

		case 'e':
			depth--
			i++

			if depth == 0 {
				return data[start:i], nil
			}

		default:
			if data[i] >= '0' && data[i] <= '9' {
				colon := bytes.IndexByte(data[i:], ':')
				if colon == -1 {
					return nil, errors.New("invalid bencode string")
				}

				colon += i

				length, err := strconv.Atoi(string(data[i:colon]))
				if err != nil {
					return nil, err
				}

				i = colon + 1 + length
			} else if data[i] == 'i' {
				end := bytes.IndexByte(data[i+1:], 'e')
				if end == -1 {
					return nil, errors.New("invalid bencode integer")
				}

				i += end + 2
			} else {
				return nil, fmt.Errorf("unexpected byte %q", data[i])
			}
		}
	}

	return nil, errors.New("unterminated info dictionary")
}

func GetInfoHash(filePath string) ([20]byte, error) {
	infoData, err := ExtractInfoRaw(filePath)
	if err != nil {
		return [20]byte{}, err
	}

	return sha1.Sum(infoData), nil
}

type TorrentState struct {
	TorrentRef Torrent
	InfoHash   [20]byte
	PeerId     [20]byte
	Key        uint32
	Downloaded uint64
	Left       uint64
	Uploaded   uint64
}

func InitializeTorrentState(infoHash [20]byte, torrent Torrent) TorrentState {
	var torrentSize uint64 = uint64(torrent.Info.PieceLength) * uint64(len(torrent.Info.Pieces))

	return TorrentState{
		InfoHash: infoHash,
		PeerId:   utils.GeneratePeerId(),
		Key:      utils.GenerateKey(),
		Left:     torrentSize, // @todo! Add disk checking for stats
	}
}
