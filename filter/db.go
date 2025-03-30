package filter

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Инициализация базы данных
func InitDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening database: %v", err)
	}
	// Создание таблицы для хранения IP-адресов TOR
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tor_nodes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ip TEXT UNIQUE
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Error creating table: %v", err)
	}

	return db, nil
}

// Вставка IP-адреса в базу данных
func InsertIP(db *sql.DB, ip string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO tor_nodes (ip) VALUES (?)", ip)
	return err
}

// Получение всех IP-адресов
func GetAllIPs(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT ip FROM tor_nodes")
	if err != nil {
		return nil, fmt.Errorf("Error fetching IPs: %v", err)
	}
	defer rows.Close()
	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, fmt.Errorf("Error scanning IP: %v", err)
		}
		ips = append(ips, ip)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating rows: %v", err)
	}
	return ips, nil
}

// Экспорт IP-адресов в файл для использования с xtables-addons
func ExportToGeoIPFile(db *sql.DB, filename string) error {
	ips, err := GetAllIPs(db)
	if err != nil {
		return fmt.Errorf("Error fetching IPs: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()

	for _, ip := range ips {
		_, err := file.WriteString(ip + "\n")
		if err != nil {
			return fmt.Errorf("Error writing IP to file: %v", err)
		}
	}

	return nil
}
