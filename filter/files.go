package filter

import (
    "bufio"
    "fmt"
    "os"

    "tor-filtering/logger"
)

type IPDiff struct{}

func NewIPDiff() *IPDiff {
    return &IPDiff{}
}

func (d *IPDiff) Diff(prev, curr []string) (remove, add []string) {
    prevMap := make(map[string]struct{}, len(prev))
    for _, ip := range prev {
        prevMap[ip] = struct{}{}
    }
    currMap := make(map[string]struct{}, len(curr))
    for _, ip := range curr {
        currMap[ip] = struct{}{}
    }
    for ip := range prevMap {
        if _, ok := currMap[ip]; !ok {
            remove = append(remove, ip)
        }
    }
    for ip := range currMap {
        if _, ok := prevMap[ip]; !ok {
            add = append(add, ip)
        }
    }
    return
}

type FileManager struct {
    Logger logger.Logger
}

func NewFileManager(l logger.Logger) *FileManager {
    return &FileManager{Logger: l}
}

func (fm *FileManager) WriteIPsToFile(path string, ips []string) error {
    file, err := os.Create(path)
    if err != nil {
        fm.Logger.Logf("error creating file %s: %v", path, err)
        return fmt.Errorf("error creating file %s: %v", path, err)
    }
    defer file.Close()
    w := bufio.NewWriter(file)
    for _, ip := range ips {
        if _, err := w.WriteString(ip + "\n"); err != nil {
            fm.Logger.Logf("error writing IP %s: %v", ip, err)
            return fmt.Errorf("error writing IP %s: %v", ip, err)
        }
    }
    if err := w.Flush(); err != nil {
        fm.Logger.Logf("error flushing file %s: %v", path, err)
        return fmt.Errorf("error flushing file %s: %v", path, err)
    }
    return nil
}

func (fm *FileManager) ReadIPsFromFile(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        fm.Logger.Logf("error opening file %s: %v", path, err)
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
        fm.Logger.Logf("error scanning file %s: %v", path, err)
        return nil, fmt.Errorf("error scanning file %s: %v", path, err)
    }
    return ips, nil
}
