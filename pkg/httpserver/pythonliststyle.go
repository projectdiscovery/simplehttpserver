package httpserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const (
	preTag      = "<pre>"
	preTagClose = "</pre>"
	aTag        = "<a"
	htmlHeader  = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Directory listing for %s</title>
</head>
<body>
`
	htmlFooter = `<hr>
</body>
</html>
`
)

type pythonStyleHandler struct {
	origWriter http.ResponseWriter
	root       http.FileSystem
}

func (h *pythonStyleHandler) Header() http.Header {
	return h.origWriter.Header()
}

func (h *pythonStyleHandler) writeListItem(b []byte, written *int) {
	var i int
	i, _ = fmt.Fprint(h.origWriter, "<li>")
	*written += i
	i, _ = h.origWriter.Write(bytes.Trim(b, "\r\n"))
	*written += i
	i, _ = fmt.Fprint(h.origWriter, "</li>\n")
	*written += i
}

func (h *pythonStyleHandler) Write(b []byte) (int, error) {
	var i int
	written := 0

	if bytes.HasPrefix(b, []byte(preTag)) {
		_, _ = io.Discard.Write(b)
		i, _ = fmt.Fprintln(h.origWriter, "<ul>")
		written += i
		return written, nil
	}
	if bytes.HasPrefix(b, []byte(preTagClose)) {
		_, _ = io.Discard.Write(b)
		i, _ = fmt.Fprintln(h.origWriter, "</ul>")
		written += i
		return written, nil
	}

	if bytes.HasPrefix(b, []byte(aTag)) {
		h.writeListItem(b, &written)
	}
	return i, nil
}

func (h *pythonStyleHandler) WriteHeader(statusCode int) {
	h.origWriter.WriteHeader(statusCode)
}

func (h *pythonStyleHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Use the FileSystem's Open to respect sandbox containment checks.
	// This avoids building raw OS paths and correctly delegates to SandboxFileSystem
	// or any other http.FileSystem implementation.
	f, err := h.root.Open(request.URL.Path)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil || !fi.IsDir() {
		// Not a directory: serve through the FileSystem handler
		http.FileServer(h.root).ServeHTTP(writer, request)
		return
	}
	_, _ = fmt.Fprintf(writer, htmlHeader, request.URL.Path)
	_, _ = fmt.Fprintf(writer, "<h1>Directory listing for %s</h1>\n<hr>\n", request.URL.Path)
	h.origWriter = writer
	http.FileServer(h.root).ServeHTTP(h, request)
	_, _ = fmt.Fprint(writer, htmlFooter)
}

// PythonStyle returns an HTTP handler that renders directory listings in Python's
// SimpleHTTPServer style.  It accepts any http.FileSystem so that sandbox
// containment (SandboxFileSystem) is honoured even in this mode.
func PythonStyle(root http.FileSystem) http.Handler {
	return &pythonStyleHandler{
		root: root,
	}
}
