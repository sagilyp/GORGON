package filter

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	ExitURL  = "https://1275.ru/list/tor_exit.txt"
	GuardURL = "https://1275.ru/list/tor_guards.csv"
)

func isIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

func FetchExitNodes() ([]string, error) {
	resp, err := http.Get(ExitURL)
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
	resp, err := http.Get(GuardURL)
	if err != nil {
		return nil, fmt.Errorf("GET guard list %s: %w", GuardURL, err)
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	var out []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse guard CSV: %w", err)
		}
		if len(record) >= 2 {
			ip := strings.TrimSpace(record[1])
			if ip != "" && isIPv4(ip) {
				out = append(out, ip)
			}
		}
	}
	return out, nil
}
