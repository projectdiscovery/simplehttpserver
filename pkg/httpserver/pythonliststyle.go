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
	origWriter   http.ResponseWriter
	goFileServer http.Handler
}

func (h pythonStyleHandler) Header() http.Header {
	return h.origWriter.Header()
}

func (h pythonStyleHandler) writeListItem(b []byte, written *int) {
	var i int
	i, _ = fmt.Fprint(h.origWriter, "<li>")
	*written += i
	i, _ = h.origWriter.Write(bytes.Trim(b, "\r\n"))
	*written += i
	i, _ = fmt.Fprint(h.origWriter, "</li>\n")
	*written += i
}

func (h pythonStyleHandler) Write(b []byte) (int, error) {
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

func (h pythonStyleHandler) WriteHeader(statusCode int) {
	h.origWriter.WriteHeader(statusCode)
}

func (h pythonStyleHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_, _ = fmt.Fprintf(writer, htmlHeader, request.URL.Path)
	_, _ = fmt.Fprintf(writer, "<h1>Directory listing for %s</h1>\n<hr>\n", request.URL.Path)

	h.origWriter = writer
	h.goFileServer.ServeHTTP(h, request)

	_, _ = fmt.Fprint(writer, htmlFooter)
}

func PythonStyle(handler http.Handler) http.Handler {
	return pythonStyleHandler{
		goFileServer: handler,
	}
}