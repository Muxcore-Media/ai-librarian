package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	manifest "github.com/Muxcore-Media/ai-librarian"
	"github.com/Muxcore-Media/core/pkg/contracts"
	modulesdk "github.com/Muxcore-Media/core/sdk/go/module"
	"google.golang.org/grpc"
)

type Module struct {
	id, grpcAddr, httpAddr string
	mu                     sync.RWMutex
	last                   []Finding
	grpcSrv                *grpc.Server
	lis                    net.Listener
	httpSrv                *http.Server
}

func NewModule() *Module { return New(Config{}) }

type Config struct{ ID, GRPCAddr, HTTPAddr string }

func New(cfg Config) *Module {
	if cfg.ID == "" {
		cfg.ID = "ai-librarian"
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = "127.0.0.1:9770"
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "127.0.0.1:9771"
	}
	if v := os.Getenv("AI_LIBRARIAN_GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := os.Getenv("MUXCORE_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	return &Module{id: cfg.ID, grpcAddr: cfg.GRPCAddr, httpAddr: cfg.HTTPAddr}
}

func (m *Module) Info() contracts.ModuleInfo {
	return contracts.ModuleInfo{
		ID: m.id, Name: "AI Librarian", Version: modulesdk.ManifestVersion(manifest.ManifestJSON),
		Roles: []string{"ai"}, Description: "AI library QA: unidentified files, wrong matches, collection completeness",
		Author: "Muxcore-Media", Capabilities: []string{"ai.librarian", "settings"},
		MinCoreVersion: MinCoreVersion, HTTPAddr: m.grpcAddr,
	}
}

func (m *Module) Init(context.Context) error { return nil }
func (m *Module) Start(ctx context.Context) error {
	return startHTTPGRPC(ctx, m.id, &m.grpcAddr, &m.httpAddr, &m.lis, &m.grpcSrv, &m.httpSrv, m, m.routes)
}
func (m *Module) Stop(ctx context.Context) error { return stopServers(ctx, m.grpcSrv, m.httpSrv) }
func (m *Module) Health(context.Context) error   { return nil }
func (m *Module) GRPCListenAddr() string         { return m.grpcAddr }
func (m *Module) HTTPListenAddr() string         { return m.httpAddr }
func (m *Module) Settings() []contracts.SettingDef {
	return []contracts.SettingDef{{Key: "scan_on_import", Label: "Scan on import events", Type: contracts.SettingTypeBool, Value: "true", Default: "true", Group: "AI"}}
}
func (m *Module) UpdateSetting(key, value string) error {
	if key != "scan_on_import" {
		return fmt.Errorf("unknown setting %q", key)
	}
	return nil
}

func (m *Module) routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/scan", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Library []LibraryItem `json:"library"`
		}
		if !readJSON(w, r, &body) {
			return
		}
		findings := scanLibrary(body.Library)
		m.mu.Lock()
		m.last = findings
		m.mu.Unlock()
		writeJSON(w, map[string]any{"findings": findings, "count": len(findings)})
	})
	mux.HandleFunc("GET /v1/findings", func(w http.ResponseWriter, _ *http.Request) {
		m.mu.RLock()
		defer m.mu.RUnlock()
		writeJSON(w, map[string]any{"findings": m.last, "count": len(m.last)})
	})
}

func (m *Module) handleMesh(_ context.Context, method string, payload []byte) ([]byte, error) {
	if method != "Scan" {
		return nil, fmt.Errorf("unknown method %s", method)
	}
	var body struct {
		Library []LibraryItem `json:"library"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, err
	}
	findings := scanLibrary(body.Library)
	m.mu.Lock()
	m.last = findings
	m.mu.Unlock()
	return json.Marshal(map[string]any{"findings": findings, "count": len(findings)})
}
