package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

const (
	TelegramBotToken = "7779836915:AAGZJ8BaJ6se0ryjW9_KHL3INBLi8RGueRo"
	TelegramChatID   = "-4787521880"
	matchPrefix      = "TOR_BLOCK"
	httpTimeout      = 5 * time.Second
)

func StartJournalAlerts(ctx context.Context) {
	cmd := exec.CommandContext(ctx, "journalctl", "-kf", "--output=cat")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("alerter: stdout pipe error: %v\n", err)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("alerter: failed to start journalctl: %v\n", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, matchPrefix) {
			go sendTelegramAlert(line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("alerter: scanner error: %v\n", err)
	}
	cmd.Wait()
}

func sendTelegramAlert(msg string) {
	escaped := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(msg)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramBotToken)
	if len(escaped) > 300 {
		escaped = escaped[:300] + "…"
	}
	text := fmt.Sprintf("🚨 <b>Blocked TOR</b>:\n<code>%s</code>", escaped)

	form := url.Values{
		"chat_id":    {TelegramChatID},
		"text":       {text},
		"parse_mode": {"HTML"},
	}

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.PostForm(apiURL, form)
	if err != nil {
		log.Printf("alerter: Telegram send error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("alerter: Telegram API returned %s", resp.Status)
	}
}
