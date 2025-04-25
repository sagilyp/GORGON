package filter

import (
	"fmt"
	"os/exec"
)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %v: %s", name, args, err, string(out))
	}
	return nil
}

func applyLogRule(chain, set, match string) error {
	prefix := fmt.Sprintf("TOR_BLOCK %s: ", set)
	if err := runCmd("sudo", "iptables", "-C", chain,
		"-m", "set", "--match-set", set, match,
		"-j", "LOG", "--log-prefix", prefix, "--log-level", "4",
	); err == nil {
		return nil
	}
	return runCmd("sudo", "iptables", "-I", chain, "1",
		"-m", "set", "--match-set", set, match,
		"-j", "LOG", "--log-prefix", prefix, "--log-level", "4",
	)
}

func RemoveIPsFromIPSet(setName string, ips []string) error {
	for _, ip := range ips {
		runCmd("sudo", "ipset", "del", setName, ip)
	}
	return nil
}

func ClearIPSet(setName string) {
	runCmd("sudo", "ipset", "flush", setName)
}

func AddIPsToIPSet(setName string, ips []string) error {
	runCmd("sudo", "ipset", "create", setName, "hash:ip", "-exist")
	for _, ip := range ips {
		if err := runCmd("sudo", "ipset", "add", setName, ip, "-exist"); err != nil {
			return fmt.Errorf("add %s to %s: %w", ip, setName, err)
		}
	}
	return nil
}

func applyRule(chain, setName, matchDir string) error {
	if err := runCmd("sudo", "iptables", "-C", chain, "-m", "set", "--match-set", setName, matchDir, "-j", "DROP"); err == nil {
		return nil
	}
	return runCmd("sudo", "iptables", "-A", chain, "-m", "set", "--match-set", setName, matchDir, "-j", "DROP")
}

func ApplyInbound(setName string) error {
	if err := applyLogRule("INPUT", setName, "src"); err != nil {
		return fmt.Errorf("apply inbound log %s: %w", setName, err)
	}
	return applyRule("INPUT", setName, "src")
}

func ApplyOutbound(setName string) error {
	if err := applyLogRule("OUTPUT", setName, "dst"); err != nil {
		return fmt.Errorf("apply outbound log %s: %w", setName, err)
	}
	return applyRule("OUTPUT", setName, "dst")
}
