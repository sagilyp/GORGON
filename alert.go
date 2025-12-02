
package main

import (
       "bufio"
       "context"
       "fmt"
       "net/http"
       "net/url"
       "os/exec"
       "strings"
       "time"
       "tor-filtering/logger"
)

type TelegramAlerter struct {
       Logger         logger.Logger
       BotToken       string
       ChatID         string
       MatchPrefix    string
       Timeout        time.Duration
}

func NewTelegramAlerter(l logger.Logger, token, chatID, matchPrefix string, timeout time.Duration) *TelegramAlerter {
       return &TelegramAlerter{
	       Logger: l,
	       BotToken: token,
	       ChatID: chatID,
	       MatchPrefix: matchPrefix,
	       Timeout: timeout,
       }
}

func (t *TelegramAlerter) StartJournalAlerts(ctx context.Context) {
       cmd := exec.CommandContext(ctx, "journalctl", "-kf", "--output=cat")
       stdout, err := cmd.StdoutPipe()
       if err != nil {
	       t.Logger.Logf("alerter: stdout pipe error: %v", err)
	       return
       }
       if err := cmd.Start(); err != nil {
	       t.Logger.Logf("alerter: failed to start journalctl: %v", err)
	       return
       }

       scanner := bufio.NewScanner(stdout)
       for scanner.Scan() {
	       line := scanner.Text()
	       if strings.Contains(line, t.MatchPrefix) {
		       go t.SendTelegramAlert(line)
	       }
       }
       if err := scanner.Err(); err != nil {
	       t.Logger.Logf("alerter: scanner error: %v", err)
       }
       cmd.Wait()
}

func (t *TelegramAlerter) SendTelegramAlert(msg string) {
       escaped := strings.NewReplacer(
	       "&", "&amp;",
	       "<", "&lt;",
	       ">", "&gt;",
       ).Replace(msg)

       apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)
       if len(escaped) > 300 {
	       escaped = escaped[:300] + "\u2026"
       }
       text := fmt.Sprintf("\U0001f6a8 <b>Blocked TOR</b>:\n<code>%s</code>", escaped)

       form := url.Values{
	       "chat_id":    {t.ChatID},
	       "text":       {text},
	       "parse_mode": {"HTML"},
       }

       client := &http.Client{Timeout: t.Timeout}
       resp, err := client.PostForm(apiURL, form)
       if err != nil {
	       t.Logger.Logf("alerter: Telegram send error: %v", err)
	       return
       }
       defer resp.Body.Close()
       if resp.StatusCode != http.StatusOK {
	       t.Logger.Logf("alerter: Telegram API returned %s", resp.Status)
       }
}
