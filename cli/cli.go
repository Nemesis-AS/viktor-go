package cli

import (
	"errors"
	"os"
	"path"
)

func ParseArgs() (string, error) {
	args := os.Args[1:]

	if len(args) == 0 {
		return "", errors.New("No torrent file specified")
	}

	torrentPath := args[0]

	cwd, _ := os.Getwd()
	filePath := path.Join(cwd, torrentPath)

	return filePath, nil
}
