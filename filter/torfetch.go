package filter

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// Временная заглушка для обновления базы данных TOR-узлов
func UpdateTorDB(database *sql.DB) error {
	data, err := os.ReadFile("torbulkexitlist") // Читаем файл с IP
	if err != nil {
		return fmt.Errorf("Ошибка чтения файла: %v", err)
	}

	ipList := strings.Split(string(data), "\n") // Разбиваем по строкам

	_, err = database.Exec("DELETE FROM tor_nodes") // Очищаем таблицу
	if err != nil {
		return fmt.Errorf("Ошибка очистки БД: %v", err)
	}

	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			_, err = database.Exec("INSERT INTO tor_nodes (ip) VALUES (?)", ip)
			if err != nil {
				return fmt.Errorf("Ошибка вставки IP %s: %v", ip, err)
			}
		}
	}

	fmt.Println("БД TOR обновлена из файла torbulkexitlist.")
	return nil
}

// // Обновление базы данных с выходными узлами TOR
// func UpdateTorDB(database *sql.DB) error {
// 	// Получаем список IP-адресов выходных узлов TOR
// 	resp, err := http.Get("https://check.torproject.org/torbulkexitlist")
// 	if err != nil {
// 		return fmt.Errorf("Error fetching TOR IP list: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return fmt.Errorf("Error reading response body: %v", err)
// 	}

// 	lines := strings.Split(string(body), "\n")
// 	for _, line := range lines {
// 		if line != "" {
// 			err := InsertIP(database, line)
// 			if err != nil {
// 				fmt.Printf("Failed to insert IP %s: %v\n", line, err)
// 			}
// 		}
// 	}

// 	return nil
// }
