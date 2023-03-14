package test

import (
	"github.com/projectdiscovery/simplehttpserver/pkg/httpserver"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestWritePythonStyleHtmlPage(t *testing.T) {
	py := httpserver.PythonStyle(http.FileServer(http.Dir("./fixture/pythonliststyle")))

	w := httptest.NewRecorder()
	py.ServeHTTP(w, httptest.NewRequest("GET", "http://example.com/", nil))
	b, _ := io.ReadAll(w.Result().Body)

	body := string(b)
	if strings.Compare(want, body) != 0 {
		t.Errorf("want:\n%s\ngot:\n%s", want, body)
	}
}
