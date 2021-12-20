package httpserver

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/simplehttpserver/pkg/unit"
)

// uploadlayer handles PUT requests and save the file to disk
func (t *HTTPServer) uploadlayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handles file write if enabled
		if EnableUpload && r.Method == http.MethodPut {
			// sandbox - calcolate absolute path
			if t.options.Sandbox {
				absPath, err := filepath.Abs(filepath.Join(t.options.Folder, r.URL.Path))
				if err != nil {
					gologger.Print().Msgf("%s\n", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// check if the path is within the configured folder
				pattern := t.options.Folder + string(filepath.Separator) + "*"
				matched, err := filepath.Match(pattern, absPath)
				if err != nil {
					gologger.Print().Msgf("%s\n", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				} else if !matched {
					gologger.Print().Msg("pointing to unauthorized directory")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			var (
				data []byte
				err  error
			)
			if t.options.Sandbox {
				maxFileSize := unit.ToMb(t.options.MaxFileSize)
				// check header content length
				if r.ContentLength > maxFileSize {
					gologger.Print().Msg("request too large")
					return
				}
				// body max length
				r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
			}

			data, err = ioutil.ReadAll(r.Body)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			sanitizedPath := filepath.FromSlash(path.Clean("/" + strings.Trim(r.URL.Path, "/")))

			err = handleUpload(t.options.Folder, sanitizedPath, data)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusCreated)
				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}

func handleUpload(base, file string, data []byte) error {
	// rejects all paths containing a non exhaustive list of invalid characters - This is only a best effort as the tool is meant for development
	if strings.ContainsAny(file, "\\`\"':") {
		return errors.New("invalid character")
	}

	untrustedPath := filepath.Clean(filepath.Join(base, file))
	if !strings.HasPrefix(untrustedPath, filepath.Clean(base)) {
		return errors.New("invalid path")
	}
	trustedPath := untrustedPath

	if _, err := os.Stat(path.Dir(trustedPath)); os.IsNotExist(err) {
		return errors.New("invalid path")
	}

	return ioutil.WriteFile(trustedPath, data, 0655)
}
