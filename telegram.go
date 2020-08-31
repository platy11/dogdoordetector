/*
	Copyright (c) 2020 platy11.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, version 3.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TgSendMessage contacts the Telegram Bot API and sends a specified message to
// a specified chat
func TgSendMessage(botToken string, chatID string, message string, silent bool) error {
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", message)
	params.Set("parse_mode", "MarkdownV2")
	params.Set("disable_notification", strconv.FormatBool(silent))
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?%s", botToken, params.Encode())

	var tries uint
	backoff := 2 * time.Second
	for {
		tries++
		resp, err := http.Get(url)
		if err != nil && tries >= 5 {
			return fmt.Errorf("HTTP GET Telegram API sendMessage: max tries reached: %w", err)
		} else if err != nil {
			log.Println("Could not HTTP GET Telegram API: retrying")
			backoff *= 2
			time.Sleep(backoff)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// TODO: Could backoff-retry for temporary server errors
			return fmt.Errorf("HTTP GET Telegram API sendMessage: HTTP error response code %v", resp.StatusCode)
		}
		break
	}
	return nil
}
