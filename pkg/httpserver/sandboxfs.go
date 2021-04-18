package httpserver

import (
	"errors"
	"net/http"
	"path/filepath"
)

// SandboxFileSystem implements superbasic security checks
type SandboxFileSystem struct {
	fs         http.FileSystem
	RootFolder string
}

// Open performs basic security checks before providing folder/file content
func (sbfs SandboxFileSystem) Open(path string) (http.File, error) {
	abspath, err := filepath.Abs(filepath.Join(sbfs.RootFolder, path))
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(abspath)
	// rejects names starting with a dot like .file
	dotmatch, err := filepath.Match(".*", filename)
	if err != nil {
		return nil, err
	} else if dotmatch {
		return nil, errors.New("invalid file")
	}

	// reject symlinks
	symlinkCheck, err := filepath.EvalSymlinks(abspath)
	if err != nil {
		return nil, err
	}
	if symlinkCheck != abspath {
		return nil, errors.New("symlinks not allowed")
	}

	// check if the path is within the configured folder
	if sbfs.RootFolder != abspath {
		pattern := sbfs.RootFolder + string(filepath.Separator) + "*"
		matched, err := filepath.Match(pattern, abspath)
		if err != nil {
			return nil, err
		} else if !matched {
			return nil, errors.New("invalid file")
		}
	}

	f, err := sbfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}
