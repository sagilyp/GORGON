package main

import (
	"context"
	"time"
    "os"
	"bufio"
	"strings"
    "os/signal"
    "syscall"
	"tor-filtering/filter"
	"tor-filtering/logger"
)

const (
	ExitSet        = "tor_exit_ips"
	GuardSet       = "tor_guard_ips"
	BridgeSet      = "tor_bridge_ips"
	GuardTxt       = "ips/guard_nodes.txt"
	ExitTxt        = "ips/exit_nodes.txt"
	BridgeTxt      = "ips/bridge_nodes.txt"
	UpdateInterval = 6 * time.Hour
)

func updateAll(
	nodeMgr *filter.TorNodeManager,
	xdpMgr *filter.XDPManager,
	tcEgressMgr *filter.TCManager,
	fileMgr *filter.FileManager,
	ipDiff *filter.IPDiff,
) {
	nodeMgr.Logger.Logf("Updating TOR lists ...")

	// Exit nodes
	prevExit, _ := fileMgr.ReadIPsFromFile(ExitTxt)
	newExit, err := nodeMgr.FetchExitNodes()
	if err != nil {
		nodeMgr.Logger.Logf("FetchExitNodes error: %v. Loading cached IPs.", err)
		xdpMgr.AddIPsToIPSet(ExitSet, prevExit)
	} else {
		toRemove, toAdd := ipDiff.Diff(prevExit, newExit)
		xdpMgr.RemoveIPsFromIPSet(ExitSet, toRemove)
		xdpMgr.AddIPsToIPSet(ExitSet, toAdd)
		_ = fileMgr.WriteIPsToFile(ExitTxt, newExit)
	}

	// Guard nodes
	prevGuard, _ := fileMgr.ReadIPsFromFile(GuardTxt)
	newGuard, err := nodeMgr.FetchGuardNodes()
	if err != nil {
		nodeMgr.Logger.Logf("FetchGuardNodes error: %v. Loading cached IPs.", err)
		tcEgressMgr.AddGuardIPs(GuardSet, prevGuard)
	} else {
		toRemove, toAdd := ipDiff.Diff(prevGuard, newGuard)
		tcEgressMgr.RemoveGuardIPs(GuardSet, toRemove)
		tcEgressMgr.AddGuardIPs(GuardSet, toAdd)
		_ = fileMgr.WriteIPsToFile(GuardTxt, newGuard)
	}

	// Bridges
	bridges, err := fileMgr.ReadIPsFromFile(BridgeTxt)
	if err != nil {
		nodeMgr.Logger.Logf("ReadLines bridges: %v", err)
	} else {
		tcEgressMgr.ClearBridgeIPs(BridgeSet)
		tcEgressMgr.AddBridgeIPs(BridgeSet, bridges)
	}

	nodeMgr.Logger.Logf("=== updating completed ===")
}

func TracePipeToFastLog(ctx context.Context, logImpl logger.Logger, prefix string) {
	file, err := os.Open("/sys/kernel/debug/tracing/trace_pipe")
	if err != nil {
		logImpl.Logf("Failed to open trace_pipe: %v (did you mount debugfs?)", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, prefix) {
			idx := strings.Index(line, prefix)
			logImpl.Logf("%s", line[idx:])
		}
	}
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    logImpl := &logger.LoggerImpl{}
    fileMgr := filter.NewFileManager(logImpl)
    nodeMgr := filter.NewTorNodeManager(
        logImpl,
        "https://check.torproject.org/torbulkexitlist",
        "/var/lib/tor/cached-microdesc-consensus",
        "127.0.0.1:9050",
    )
    ipDiff := filter.NewIPDiff()
    
    xdpMgr, err := filter.NewXDPManager(logImpl)
    if err != nil {
        logImpl.Logf("FATAL: init XDP ingress: %v", err)
        return
    }
    defer xdpMgr.Close()

    tcEgressMgr, err := filter.NewTCManager(logImpl, filter.TCDirectionEgress)
    if err != nil {
        logImpl.Logf("FATAL: init TC egress: %v", err)
        return
    }
    defer tcEgressMgr.Close()
	tgToken := os.Getenv("TG_BOT_TOKEN")
    tgChatID := os.Getenv("TG_CHAT_ID")
	if tgToken == "" || tgChatID == "" {
        logImpl.Logf("FATAL: Не заданы переменные окружения TG_BOT_TOKEN или TG_CHAT_ID")
        return
    }
    telegramAlerter := NewTelegramAlerter(
        logImpl,
        tgToken,
        tgChatID,
        "TOR_DETECT",
        5*time.Second,
    )
    go telegramAlerter.StartJournalAlerts(ctx)
    go TracePipeToFastLog(ctx, logImpl, "TOR_DETECT")
    updateAll(nodeMgr, xdpMgr, tcEgressMgr, fileMgr, ipDiff)
    ticker := time.NewTicker(UpdateInterval)
    defer ticker.Stop()
    for range ticker.C {
        updateAll(nodeMgr, xdpMgr, tcEgressMgr, fileMgr, ipDiff)
    }
    logImpl.Logf("Gorgon IDS started, waiting for signals...")
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    <-sigCh 
    logImpl.Logf("Shutting down... detaching eBPF filters from interfaces...")
}
