package httpserver

import (
	"net/http"
	"testing"
)

// TestPythonStyleAcceptFileSystem verifies that PythonStyle accepts an
// http.FileSystem interface (not a concrete http.Dir).  Before the fix,
// httpserver.New() panicked with
//
//	interface conversion: http.FileSystem is httpserver.SandboxFileSystem, not http.Dir
//
// when both -sandbox and -python flags were set, because PythonStyle only
// accepted http.Dir via a type assertion.
func TestPythonStyleAcceptFileSystem(t *testing.T) {
	// SandboxFileSystem implements http.FileSystem but is NOT http.Dir.
	// Passing it to PythonStyle must not panic.
	sbfs := SandboxFileSystem{
		fs:         http.Dir(t.TempDir()),
		RootFolder: t.TempDir(),
	}
	// This call would panic on the old code because of dir.(http.Dir).
	handler := PythonStyle(sbfs)
	if handler == nil {
		t.Fatal("PythonStyle returned nil handler")
	}
}

// TestNewServerSandboxPythonNoPanic mirrors the exact code path in New()
// that used to panic: Sandbox=true + Python=true.
func TestNewServerSandboxPythonNoPanic(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		Folder:        dir,
		ListenAddress: "127.0.0.1:0",
		Sandbox:       true,
		Python:        true,
	}
	// Before the fix this returned a panic, not an error.
	srv, err := New(opts)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("New() returned nil server")
	}
}
