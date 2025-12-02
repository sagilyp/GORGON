
package filter

import (
       "fmt"
       "os/exec"
       "tor-filtering/logger"
)

type IPTablesManager struct {
       Logger logger.Logger
}

func NewIPTablesManager(l logger.Logger) *IPTablesManager {
       return &IPTablesManager{Logger: l}
}

func (m *IPTablesManager) runCmd(name string, args ...string) error {
       cmd := exec.Command(name, args...)
       if out, err := cmd.CombinedOutput(); err != nil {
	       m.Logger.Logf("runCmd error: %s %v: %v: %s", name, args, err, string(out))
	       return fmt.Errorf("%s %v: %v: %s", name, args, err, string(out))
       }
       return nil
}

func (m *IPTablesManager) applyLogRule(chain, set, match string) error {
       prefix := fmt.Sprintf("TOR_BLOCK %s: ", set)
       if err := m.runCmd("sudo", "iptables", "-C", chain,
	       "-m", "set", "--match-set", set, match,
	       "-j", "LOG", "--log-prefix", prefix, "--log-level", "4",
       ); err == nil {
	       return nil
       }
       return m.runCmd("sudo", "iptables", "-I", chain, "1",
	       "-m", "set", "--match-set", set, match,
	       "-j", "LOG", "--log-prefix", prefix, "--log-level", "4",
       )
}

func (m *IPTablesManager) RemoveIPsFromIPSet(setName string, ips []string) error {
       for _, ip := range ips {
	       m.runCmd("sudo", "ipset", "del", setName, ip)
       }
       return nil
}

func (m *IPTablesManager) ClearIPSet(setName string) {
       m.runCmd("sudo", "ipset", "flush", setName)
}

func (m *IPTablesManager) AddIPsToIPSet(setName string, ips []string) error {
       m.runCmd("sudo", "ipset", "create", setName, "hash:ip", "-exist")
       for _, ip := range ips {
	       if err := m.runCmd("sudo", "ipset", "add", setName, ip, "-exist"); err != nil {
		       return fmt.Errorf("add %s to %s: %w", ip, setName, err)
	       }
       }
       return nil
}

func (m *IPTablesManager) applyRule(chain, setName, matchDir string) error {
       if err := m.runCmd("sudo", "iptables", "-C", chain, "-m", "set", "--match-set", setName, matchDir, "-j", "DROP"); err == nil {
	       return nil
       }
       return m.runCmd("sudo", "iptables", "-A", chain, "-m", "set", "--match-set", setName, matchDir, "-j", "DROP")
}

func (m *IPTablesManager) ApplyInbound(setName string) error {
       if err := m.applyLogRule("INPUT", setName, "src"); err != nil {
	       return fmt.Errorf("apply inbound log %s: %w", setName, err)
       }
       return m.applyRule("INPUT", setName, "src")
}

func (m *IPTablesManager) ApplyOutbound(setName string) error {
       if err := m.applyLogRule("FORWARD", setName, "dst"); err != nil {
	       return fmt.Errorf("apply outbound log %s: %w", setName, err)
       }
       return m.applyRule("FORWARD", setName, "dst")
}
