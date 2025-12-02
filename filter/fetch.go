
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
       "tor-filtering/logger"
)

type TorNodeManager struct {
       Logger logger.Logger
       ExitURL      string
       GuardPath    string
       TorSocksAddr string
}

func NewTorNodeManager(l logger.Logger, exitURL, guardPath, socksAddr string) *TorNodeManager {
       return &TorNodeManager{
	       Logger: l,
	       ExitURL: exitURL,
	       GuardPath: guardPath,
	       TorSocksAddr: socksAddr,
       }
}

func (m *TorNodeManager) isIPv4(s string) bool {
       ip := net.ParseIP(s)
       return ip != nil && ip.To4() != nil
}

func (m *TorNodeManager) FetchExitNodes() ([]string, error) {
       dialer, err := proxy.SOCKS5("tcp", m.TorSocksAddr, nil, proxy.Direct)
       if err != nil {
	       m.Logger.Logf("SOCKS5 dialer error: %v", err)
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
       resp, err := client.Get(m.ExitURL)
       if err != nil {
	       m.Logger.Logf("GET exit list error: %v", err)
	       return nil, fmt.Errorf("GET exit list %s: %w", m.ExitURL, err)
       }
       defer resp.Body.Close()
       var out []string
       scanner := bufio.NewScanner(resp.Body)
       for scanner.Scan() {
	       ip := strings.TrimSpace(scanner.Text())
	       if ip != "" && m.isIPv4(ip) {
		       out = append(out, ip)
	       }
       }
       if err := scanner.Err(); err != nil {
	       m.Logger.Logf("scan exit list error: %v", err)
	       return nil, fmt.Errorf("scan exit list: %w", err)
       }
       return out, nil
}

func (m *TorNodeManager) FetchGuardNodes() ([]string, error) {
       f, err := os.Open(m.GuardPath)
       if err != nil {
	       m.Logger.Logf("open consensus error: %v", err)
	       return nil, fmt.Errorf("open %s: %w", m.GuardPath, err)
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
		       if m.isIPv4(lastR) {
			       out = append(out, lastR)
		       }
		       lastR = ""
	       }
       }
       if err := scanner.Err(); err != nil {
	       m.Logger.Logf("scan consensus error: %v", err)
	       return nil, fmt.Errorf("scan consensus: %w", err)
       }
       return out, nil
}
