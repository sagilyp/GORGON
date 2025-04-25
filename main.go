package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"tor-filtering/filter"
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

func updateAll() {
	fmt.Println("Updating TOR lists ...")
	// exit
	prevExit, _ := filter.ReadIPsFromFile(ExitTxt)
	newExit, err := filter.FetchExitNodes()
	if err != nil {
		log.Printf("FetchExitNodes error: %v", err)
		return
	} else {
		toRemove, toAdd := filter.DiffIPs(prevExit, newExit)
		filter.RemoveIPsFromIPSet(ExitSet, toRemove)
		filter.AddIPsToIPSet(ExitSet, toAdd)
		filter.ApplyInbound(ExitSet)
		if err := filter.WriteIPsToFile(ExitTxt, newExit); err != nil {
			log.Printf("WriteIPsToFile exit: %v", err)
		}
	}
	// guard
	prevGuard, _ := filter.ReadIPsFromFile(GuardTxt)
	newGuard, err := filter.FetchGuardNodes()
	if err != nil {
		log.Printf("FetchGuardNodes: %v", err)
	} else {
		toRemove, toAdd := filter.DiffIPs(prevGuard, newGuard)
		filter.RemoveIPsFromIPSet(GuardSet, toRemove)
		filter.AddIPsToIPSet(GuardSet, toAdd)
		filter.ApplyOutbound(GuardSet)
		if err := filter.WriteIPsToFile(GuardTxt, newGuard); err != nil {
			log.Printf("WriteIPsToFile guard: %v", err)
		}
	}
	// bridges
	bridges, err := filter.ReadIPsFromFile(BridgeTxt)
	if err != nil {
		log.Printf("ReadLines bridges: %v", err)
	} else {
		filter.ClearIPSet(BridgeSet)
		filter.AddIPsToIPSet(BridgeSet, bridges)
		filter.ApplyOutbound(BridgeSet)
	}
	fmt.Println("=== updating completed ===")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go StartJournalAlerts(ctx)

	updateAll()
	ticker := time.NewTicker(UpdateInterval)
	defer ticker.Stop()
	for range ticker.C {
		updateAll()
	}
}
