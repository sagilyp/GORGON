package main

import (
	"fmt"
	"log"
	"time"

	"tor-filtering/filter"
)

const (
	DatabaseFile   = "ips/tor_ips.db"
	TorExitListURL = "https://check.torproject.org/torbulkexitlist"
	GeoIPFilePath  = "ips/tor_nodes.csv"
	UpdateInterval = 6 * time.Hour
)

func main() {
	// Инициализация базы данных
	database, err := filter.InitDB("./tor_ips.db")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer database.Close()
	// Настройка обновления данных о выходных узлах TOR
	err = filter.UpdateTorDB(database)
	if err != nil {
		log.Fatalf("Error updating TOR IP database: %v", err)
	}
	fmt.Println("TOR exit node list updated.")
	// Экспортируем IP-адреса в файл для использования с xtables-addons
	err = filter.ExportToGeoIPFile(database, GeoIPFilePath)
	if err != nil {
		log.Fatalf("Error exporting TOR IPs to GeoIP file: %v", err)
	}
	fmt.Println("TOR IPs exported to GeoIP file.")
	// Применяем правила фильтрации с помощью iptables
	err = filter.ApplyFilteringRules()
	if err != nil {
		log.Fatalf("Error applying filtering rules: %v", err)
	}
	fmt.Println("Filtering rules applied.")
	// Настроим регулярное обновление базы данных (например, раз в 6 часов)
	ticker := time.NewTicker(6 * time.Hour)
	for {
		select {
		case <-ticker.C:
			fmt.Println("Updating TOR IP database...")
			err := filter.UpdateTorDB(database)
			if err != nil {
				log.Printf("Error updating TOR IP database: %v", err)
			} else {
				fmt.Println("TOR IP database updated.")
				err := filter.ExportToGeoIPFile(database, GeoIPFilePath)
				if err != nil {
					log.Printf("Error exporting TOR IPs: %v", err)
				}
			}
		}
	}
}
