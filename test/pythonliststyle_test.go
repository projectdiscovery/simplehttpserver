package test

import (
	"bytes"
	"github.com/projectdiscovery/simplehttpserver/pkg/httpserver"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestServePythonStyleHtmlPageForDirectories(t *testing.T) {
	const want = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title>Directory listing for /</title>
</head>
<body>
<h1>Directory listing for /</h1>
<hr>
<ul>
<li><a href="test%20file.txt">test file.txt</a></li>
</ul>
<hr>
</body>
</html>
`
	py := httpserver.PythonStyle("./fixture/pythonliststyle")

	w := httptest.NewRecorder()
	py.ServeHTTP(w, httptest.NewRequest("GET", "http://example.com/", nil))
	b, _ := io.ReadAll(w.Result().Body)

	body := string(b)
	if strings.Compare(want, body) != 0 {
		t.Errorf("want:\n%s\ngot:\n%s", want, body)
	}
}

func TestServeFileContentForFiles(t *testing.T) {
	want, _ := os.ReadFile("./fixture/pythonliststyle/test file.txt")

	py := httpserver.PythonStyle("./fixture/pythonliststyle")

	w := httptest.NewRecorder()
	py.ServeHTTP(w, httptest.NewRequest(
		"GET",
		"http://example.com/test%20file.txt",
		nil,
	))
	got, _ := io.ReadAll(w.Result().Body)
	if !bytes.Equal(want, got) {
		t.Errorf("want:\n%x\ngot:\n%x", want, got)
	}
}

func TestResponseNotFound(t *testing.T) {
	const want = `404 page not found
`

	py := httpserver.PythonStyle("./fixture/pythonliststyle")

	w := httptest.NewRecorder()
	py.ServeHTTP(w, httptest.NewRequest(
		"GET",
		"http://example.com/does-not-exist.txt",
		nil,
	))
	got, _ := io.ReadAll(w.Result().Body)
	if strings.Compare(want, string(got)) != 0 {
		t.Errorf("want:\n%s\ngot:\n%s", want, got)
	}
}
