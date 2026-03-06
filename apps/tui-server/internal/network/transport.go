package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultDNSServers = "1.1.1.1:53,8.8.8.8:53,8.8.4.4:53"

var defaultCertFiles = []string{
	"/data/data/com.termux/files/usr/etc/tls/cert.pem",
	"/etc/ssl/certs/ca-certificates.crt",
	"/etc/ssl/cert.pem",
}

// NewHTTPTransport returns an HTTP transport with a deterministic DNS fallback.
func NewHTTPTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial:     newDNSDialer(loadDNSServers()),
		},
	}
	transport.DialContext = dialer.DialContext
	transport.TLSClientConfig = &tls.Config{
		RootCAs: loadRootCAs(),
	}
	return transport
}

func newDNSDialer(servers []string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network string, _ string) (net.Conn, error) {
		var lastErr error
		baseDialer := &net.Dialer{Timeout: 5 * time.Second}
		for _, server := range servers {
			conn, err := baseDialer.DialContext(ctx, network, server)
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
}

func loadDNSServers() []string {
	value := strings.TrimSpace(os.Getenv("DNS_SERVERS"))
	if value == "" {
		value = defaultDNSServers
	}

	parts := strings.Split(value, ",")
	servers := make([]string, 0, len(parts))
	for _, part := range parts {
		server := strings.TrimSpace(part)
		if server == "" {
			continue
		}
		if _, _, err := net.SplitHostPort(server); err != nil {
			server = net.JoinHostPort(server, "53")
		}
		servers = append(servers, server)
	}

	if len(servers) == 0 {
		return []string{"1.1.1.1:53", "8.8.8.8:53", "8.8.4.4:53"}
	}

	return servers
}

func loadRootCAs() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}

	addedCerts := false
	for _, path := range certFileCandidates() {
		pemData, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		if pool.AppendCertsFromPEM(pemData) {
			addedCerts = true
		}
	}

	if addedCerts || len(pool.Subjects()) > 0 {
		return pool
	}

	return nil
}

func certFileCandidates() []string {
	seen := map[string]struct{}{}
	candidates := make([]string, 0, len(defaultCertFiles)+1)

	if value := strings.TrimSpace(os.Getenv("SSL_CERT_FILE")); value != "" {
		cleaned := filepath.Clean(value)
		seen[cleaned] = struct{}{}
		candidates = append(candidates, cleaned)
	}

	for _, path := range defaultCertFiles {
		cleaned := filepath.Clean(path)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		candidates = append(candidates, cleaned)
	}

	return candidates
}
