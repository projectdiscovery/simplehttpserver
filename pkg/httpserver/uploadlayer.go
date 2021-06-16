package httpserver

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func handleUpload(base, file string, data []byte) error {
	// rejects all paths containing a non exhaustive list of invalid characters - This is only a best effort as the tool is meant for development
	if strings.ContainsAny(file, "\\`\"':") {
		return errors.New("invalid character")
	}

	// allow upload only in subfolders
	rel, err := filepath.Rel(base, file)
	if rel == "" || err != nil {
		return err
	}

	return ioutil.WriteFile(file, data, 0655)
}
