package main

import (
	"context"
	"time"

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
    tcIngressMgr *filter.TCManager,  // ✅ Для exit nodes (входящий)
    tcEgressMgr *filter.TCManager,   // ✅ Для guard/bridge (исходящий)
    fileMgr *filter.FileManager,
    ipDiff *filter.IPDiff,
) {
    nodeMgr.Logger.Logf("Updating TOR lists ...")
    
    // Exit nodes - блокировка входящих соединений
    prevExit, _ := fileMgr.ReadIPsFromFile(ExitTxt)
    newExit, err := nodeMgr.FetchExitNodes()
    if err != nil {
        nodeMgr.Logger.Logf("FetchExitNodes error: %v", err)
    } else {
        toRemove, toAdd := ipDiff.Diff(prevExit, newExit)
        tcIngressMgr.RemoveIPsFromIPSet(ExitSet, toRemove)
        tcIngressMgr.AddIPsToIPSet(ExitSet, toAdd)
        if err := fileMgr.WriteIPsToFile(ExitTxt, newExit); err != nil {
            nodeMgr.Logger.Logf("WriteIPsToFile exit: %v", err)
        }
    }
    
    // Guard nodes - блокировка исходящих соединений
    prevGuard, _ := fileMgr.ReadIPsFromFile(GuardTxt)
    newGuard, err := nodeMgr.FetchGuardNodes()
    if err != nil {
        nodeMgr.Logger.Logf("FetchGuardNodes: %v", err)
    } else {
        toRemove, toAdd := ipDiff.Diff(prevGuard, newGuard)
        tcEgressMgr.RemoveIPsFromIPSet(GuardSet, toRemove)
        tcEgressMgr.AddIPsToIPSet(GuardSet, toAdd)
        if err := fileMgr.WriteIPsToFile(GuardTxt, newGuard); err != nil {
            nodeMgr.Logger.Logf("WriteIPsToFile guard: %v", err)
        }
    }
    
    // Bridges - блокировка исходящих соединений
    bridges, err := fileMgr.ReadIPsFromFile(BridgeTxt)
    if err != nil {
        nodeMgr.Logger.Logf("ReadLines bridges: %v", err)
    } else {
        tcEgressMgr.ClearIPSet(BridgeSet)
        tcEgressMgr.AddIPsToIPSet(BridgeSet, bridges)
    }
    
    nodeMgr.Logger.Logf("=== updating completed ===")
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
    
    // Инициализация TC ingress (для exit nodes)
    tcIngressMgr, err := filter.NewTCManager(logImpl, "eth0", filter.TCDirectionIngress)
    if err != nil {
        logImpl.Logf("FATAL: init TC ingress: %v", err)
        return
    }
    defer tcIngressMgr.Close()

    // Инициализация TC egress (для guard/bridge nodes)
    tcEgressMgr, err := filter.NewTCManager(logImpl, "eth0", filter.TCDirectionEgress)
    if err != nil {
        logImpl.Logf("FATAL: init TC egress: %v", err)
        return
    }
    defer tcEgressMgr.Close()

    telegramAlerter := NewTelegramAlerter(
        logImpl,
        "7779836915:AAGZJ8BaJ6se0ryjW9_KHL3INBLi8RGueRo",
        "-4787521880",
        "TOR_BLOCK",
        5*time.Second,
    )
    go telegramAlerter.StartJournalAlerts(ctx)

    updateAll(nodeMgr, tcIngressMgr, tcEgressMgr, fileMgr, ipDiff)
    ticker := time.NewTicker(UpdateInterval)
    defer ticker.Stop()
    for range ticker.C {
        updateAll(nodeMgr, tcIngressMgr, tcEgressMgr, fileMgr, ipDiff)
    }
}
