package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type updateTransport func(*http.Request) (*http.Response, error)

func (f updateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestUpdateFlags(t *testing.T) {
	if updateFlag := os.Getenv("SIMPLEHTTPSERVER_TEST_UPDATE_FLAG"); updateFlag != "" {
		// The shared updater exits the process, so run it in a child and
		// intercept HTTP to avoid real downloads or replacing the executable.
		t.Setenv("GITHUB_TOKEN", "")
		http.DefaultTransport = updateTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet || req.URL.String() != "https://api.github.com/repos/projectdiscovery/simplehttpserver/releases/latest" {
				t.Fatalf("unexpected update request: %s %s", req.Method, req.URL)
			}
			fmt.Println("requested simplehttpserver release")
			status := http.StatusOK
			body := fmt.Sprintf(`{"tag_name":"v%s"}`, Version)
			if os.Getenv("SIMPLEHTTPSERVER_TEST_UPDATE_ERROR") == "1" {
				status = http.StatusServiceUnavailable
				body = `{"message":"service unavailable"}`
			}
			return &http.Response{
				StatusCode: status,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		})
		os.Args = []string{"simplehttpserver", updateFlag}
		ParseOptions()
		t.Fatal("update flag returned to server startup")
	}

	for _, updateFlag := range []string{"-up", "-update"} {
		for _, fails := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/error=%t", updateFlag, fails), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestUpdateFlags$")
				errorMode := "0"
				if fails {
					errorMode = "1"
				}
				cmd.Env = append(os.Environ(),
					"SIMPLEHTTPSERVER_TEST_UPDATE_FLAG="+updateFlag,
					"SIMPLEHTTPSERVER_TEST_UPDATE_ERROR="+errorMode,
				)
				output, err := cmd.CombinedOutput()
				if ctx.Err() != nil {
					t.Fatalf("update command did not exit: %s", output)
				}
				if fails && err == nil || !fails && err != nil {
					t.Fatalf("unexpected update exit: %v\n%s", err, output)
				}
				want := "simplehttpserver is already updated to latest version"
				if fails {
					want = "failed to download latest release"
				}
				if !strings.Contains(string(output), "requested simplehttpserver release") || !strings.Contains(string(output), want) {
					t.Fatalf("missing expected updater output %q:\n%s", want, output)
				}
			})
		}
	}
}
