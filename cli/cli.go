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

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	filePath := path.Join(cwd, torrentPath)

	return filePath, nil
}
