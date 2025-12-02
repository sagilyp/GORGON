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
       iptMgr *filter.IPTablesManager,
       fileMgr *filter.FileManager,
       ipDiff *filter.IPDiff,
) {
       nodeMgr.Logger.Logf("Updating TOR lists ...")
       // exit
       prevExit, _ := fileMgr.ReadIPsFromFile(ExitTxt)
       newExit, err := nodeMgr.FetchExitNodes()
       if err != nil {
	       nodeMgr.Logger.Logf("FetchExitNodes error: %v", err)
	       return
       } else {
	       toRemove, toAdd := ipDiff.Diff(prevExit, newExit)
	       iptMgr.RemoveIPsFromIPSet(ExitSet, toRemove)
	       iptMgr.AddIPsToIPSet(ExitSet, toAdd)
	       iptMgr.ApplyInbound(ExitSet)
	       if err := fileMgr.WriteIPsToFile(ExitTxt, newExit); err != nil {
		       nodeMgr.Logger.Logf("WriteIPsToFile exit: %v", err)
	       }
       }
       // guard
       prevGuard, _ := fileMgr.ReadIPsFromFile(GuardTxt)
       newGuard, err := nodeMgr.FetchGuardNodes()
       if err != nil {
	       nodeMgr.Logger.Logf("FetchGuardNodes: %v", err)
       } else {
	       toRemove, toAdd := ipDiff.Diff(prevGuard, newGuard)
	       iptMgr.RemoveIPsFromIPSet(GuardSet, toRemove)
	       iptMgr.AddIPsToIPSet(GuardSet, toAdd)
	       iptMgr.ApplyOutbound(GuardSet)
	       if err := fileMgr.WriteIPsToFile(GuardTxt, newGuard); err != nil {
		       nodeMgr.Logger.Logf("WriteIPsToFile guard: %v", err)
	       }
       }
       // bridges
       bridges, err := fileMgr.ReadIPsFromFile(BridgeTxt)
       if err != nil {
	       nodeMgr.Logger.Logf("ReadLines bridges: %v", err)
       } else {
	       iptMgr.ClearIPSet(BridgeSet)
	       iptMgr.AddIPsToIPSet(BridgeSet, bridges)
	       iptMgr.ApplyOutbound(BridgeSet)
       }
       nodeMgr.Logger.Logf("=== updating completed ===")
}

func main() {
       ctx, cancel := context.WithCancel(context.Background())
       defer cancel()

       logImpl := &logger.LoggerImpl{}
       iptMgr := filter.NewIPTablesManager(logImpl)
       fileMgr := filter.NewFileManager(logImpl)
       nodeMgr := filter.NewTorNodeManager(logImpl, "https://check.torproject.org/torbulkexitlist", "/var/lib/tor/cached-microdesc-consensus", "127.0.0.1:9050")
       ipDiff := filter.NewIPDiff()

       telegramAlerter := NewTelegramAlerter(logImpl, "7779836915:AAGZJ8BaJ6se0ryjW9_KHL3INBLi8RGueRo", "-4787521880", "TOR_BLOCK", 5*time.Second)
       go telegramAlerter.StartJournalAlerts(ctx)

       updateAll(nodeMgr, iptMgr, fileMgr, ipDiff)
       ticker := time.NewTicker(UpdateInterval)
       defer ticker.Stop()
       for range ticker.C {
	       updateAll(nodeMgr, iptMgr, fileMgr, ipDiff)
       }
}
