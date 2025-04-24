package filter

import (
	"bufio"
	"fmt"
	"os"
)

func WriteIPsToFile(path string, ips []string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", path, err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	for _, ip := range ips {
		if _, err := w.WriteString(ip + "\n"); err != nil {
			return fmt.Errorf("error writing IP %s: %v", ip, err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing file %s: %v", path, err)
	}
	return nil
}

func ReadIPsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var ips []string
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			ips = append(ips, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file %s: %v", path, err)
	}
	return ips, nil
}
