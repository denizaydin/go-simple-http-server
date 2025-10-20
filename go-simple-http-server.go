package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type HopInfo struct {
	NodeName        string              `json:"node_name"`
	PodName         string              `json:"pod_name"`
	Hostname        string              `json:"hostname"`
	RequestSourceIP string              `json:"request_source_addr"`
	RequestDestIP   string              `json:"request_destination_addr"`
	RequestURL      string              `json:"request_url"`
	IncomingHeaders map[string][]string `json:"incoming_headers"`
	Timestamp       string              `json:"ts"`
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func detectScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		return strings.ToLower(p)
	}
	return "http"
}

func fullRequestURL(r *http.Request) string {
	return fmt.Sprintf("%s://%s%s", detectScheme(r), r.Host, r.URL.RequestURI())
}

func copyHeaders(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func buildSelfHop(r *http.Request) HopInfo {
	hostname, _ := os.Hostname()
	srcAddress := r.RemoteAddr
	dstAddress := ""

	la := r.Context().Value(http.LocalAddrContextKey)
	if la != nil {
		dstAddress = la.(net.Addr).String()
	}

	return HopInfo{
		NodeName:        getenv("NODE_NAME", ""),
		PodName:         getenv("POD_NAME", ""),
		Hostname:        hostname,
		RequestSourceIP: srcAddress,
		RequestDestIP:   dstAddress,
		RequestURL:      fullRequestURL(r),
		IncomingHeaders: copyHeaders(r.Header),
		Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func normalizeTargetURL(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("empty target url")
	}
	l := strings.ToLower(raw)
	if strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://") {
		return raw, nil
	}
	scheme := "http"
	hp := raw
	hpOnly := hp
	if i := strings.Index(hp, "/"); i >= 0 {
		hpOnly = hp[:i]
	}
	if strings.Contains(hpOnly, ":") && !strings.HasPrefix(hpOnly, "[") {
		if _, port, err := net.SplitHostPort(hpOnly); err == nil && port == "443" {
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s", scheme, raw), nil
}

func callDownstream(ctx context.Context, target string) ([]HopInfo, error) {
	u, err := normalizeTargetURL(target)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downstream call failed: %w", err)
	}
	defer resp.Body.Close()

	var chain []HopInfo
	if err := json.NewDecoder(resp.Body).Decode(&chain); err == nil {
		return chain, nil
	}
	return nil, errors.New("downstream returned non-JSON or unexpected shape")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/healthz" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	self := buildSelfHop(r)
	target := os.Getenv("CALL_SERVICE")

	if strings.TrimSpace(target) == "" {
		writeJSON(w, http.StatusOK, []HopInfo{self})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
	defer cancel()

	chain, err := callDownstream(ctx, target)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": err.Error(),
			"chain": []HopInfo{self},
		})
		return
	}
	chain = append(chain, self)
	writeJSON(w, http.StatusOK, chain)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/healthz", rootHandler)

	port := getenv("PORT", "8080")
	ipMode := strings.ToLower(getenv("IP_MODE", ""))

	addr := ":" + port
	var (
		network  string
		listener net.Listener
		err      error
	)

	switch ipMode {
	case "ipv4":
		addr = "0.0.0.0:" + port
		network = "tcp4"
	case "ipv6":
		addr = "[::]:" + port
		network = "tcp6"
	default:
		log.Printf("Starting server on %s (default stack)", addr)
		if err = http.ListenAndServe(addr, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
		return
	}

	listener, err = net.Listen(network, addr)
	if err != nil {
		log.Fatalf("failed to start %s listener on %s: %v", network, addr, err)
	}
	log.Printf("Starting server on %s with IP_MODE=%s", addr, ipMode)
	if err = http.Serve(listener, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
