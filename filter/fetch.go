package filter

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const (
	ExitURL      = "https://check.torproject.org/torbulkexitlist"
	GuardPath    = "/var/lib/tor/cached-microdesc-consensus"
	TorSocksAddr = "127.0.0.1:9050"
)

func isIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func FetchExitNodes() ([]string, error) {
	dialer, err := proxy.SOCKS5("tcp", TorSocksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("SOCKS5 dialer: %w", err)
	}
	httpTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
	}
	client := &http.Client{
		Transport: httpTransport,
		Timeout:   30 * time.Second,
	}
	resp, err := client.Get(ExitURL)
	if err != nil {
		return nil, fmt.Errorf("GET exit list %s: %w", ExitURL, err)
	}
	defer resp.Body.Close()
	var out []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" && isIPv4(ip) {
			out = append(out, ip)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan exit list: %w", err)
	}
	return out, nil
}

func FetchGuardNodes() ([]string, error) {
	f, err := os.Open(GuardPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", GuardPath, err)
	}
	defer f.Close()
	var out []string
	var lastR string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "r ") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				lastR = fields[5] // IP
			}
		}
		if strings.HasPrefix(line, "s ") && strings.Contains(line, "Guard") && lastR != "" {
			if isIPv4(lastR) {
				out = append(out, lastR)
			}
			lastR = ""
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan consensus: %w", err)
	}
	return out, nil
}
