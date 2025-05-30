package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Slackイベントの構造体
//
// SEE:
//   - https://api.slack.com/events-api
//   - https://api.slack.com/events/reaction_added
type SlackEvent struct {
	Type  string `json:"type"`
	Event struct {
		Type string `json:"type"`
		User string `json:"user"`
		Item struct {
			Type    string `json:"type"`
			Channel string `json:"channel"`
			Ts      string `json:"ts"`
		} `json:"item"`
		Reaction string `json:"reaction"`
	} `json:"event"`
	Challenge string `json:"challenge"`
}

func main() {
	http.HandleFunc("/slack/events", slackEventHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func slackEventHandler(w http.ResponseWriter, r *http.Request) {
	var event SlackEvent
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// SlackのURL検証用
	if event.Challenge != "" {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(event.Challenge)); err != nil {
			log.Println("failed to write challenge response:", err)
		}
		return
	}

	if event.Event.Type == "reaction_added" {
		go handleReactionAdded(event)
	}
	w.WriteHeader(http.StatusOK)
}

func handleReactionAdded(event SlackEvent) {
	// 1. Slack APIでメッセージテキスト取得
	text, err := fetchSlackMessageText(event.Event.Item.Channel, event.Event.Item.Ts)
	if err != nil {
		log.Println("failed to fetch message text:", err)
		return
	}
	// 2. Notionに追加（ダミー）
	if err := addToNotion(text); err != nil {
		log.Println("failed to add to Notion:", err)
	}
}

func fetchSlackMessageText(channel, ts string) (string, error) {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("SLACK_BOT_TOKEN environment variable is not set")
	}
	url := "https://slack.com/api/conversations.history?channel=" + channel + "&latest=" + ts + "&inclusive=true&limit=1"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("failed to close response body:", err)
		}
	}()
	var res struct {
		OK       bool `json:"ok"`
		Messages []struct {
			Text string `json:"text"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if !res.OK || len(res.Messages) == 0 {
		return "", nil
	}
	return res.Messages[0].Text, nil
}

func addToNotion(text string) error {
	notionToken := os.Getenv("NOTION_API_TOKEN")
	if notionToken == "" {
		return fmt.Errorf("NOTION_API_TOKEN environment variable is not set")
	}
	dbID := os.Getenv("NOTION_DB_ID")
	if dbID == "" {
		return fmt.Errorf("NOTION_DB_ID environment variable is not set")
	}

	url := "https://api.notion.com/v1/pages"
	payload := fmt.Sprintf(`{
		"parent": {"database_id": "%s"},
		"properties": {
			"テーマ": {
				"title": [
					{"text": {"content": "%s"}}
				]
			}
		}
	}`, dbID, escapeJSONString(text))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+notionToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("failed to close response body:", err)
		}
	}()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		var resBody bytes.Buffer
		if _, err := resBody.ReadFrom(resp.Body); err != nil {
			return fmt.Errorf("notion API error: failed to read response body: %w", err)
		}
		return fmt.Errorf("notion API error: %s", resBody.String())
	}
	log.Println("[Notion] テーマとして追加:", text)
	return nil
}

// JSONエスケープ用
func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, `\\`, `\\\\`)
	s = strings.ReplaceAll(s, `"`, `\\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}
