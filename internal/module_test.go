package internal

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestScanHTTP(t *testing.T) {
	m := New(Config{GRPCAddr: "127.0.0.1:0", HTTPAddr: "127.0.0.1:0"})
	if err := m.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Stop(context.Background()) })
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + m.HTTPListenAddr() + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	resp, err := http.Post("http://"+m.HTTPListenAddr()+"/v1/scan", "application/json",
		strings.NewReader(`{"library":[{"id":"1","path":"/lib/foo.mkv"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("%d %s", resp.StatusCode, raw)
	}
	if !strings.Contains(string(raw), "unidentified_file") {
		t.Fatalf("%s", raw)
	}
}
