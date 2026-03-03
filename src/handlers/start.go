/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package handlers

import (
	"fmt"
	"runtime"
	"time"

	"ashokshau/tgmusic/src/core"
	"ashokshau/tgmusic/src/core/db"

	"github.com/amarnathcjd/gogram/telegram"
)

// pingHandler handles the /ping command.
func pingHandler(m *telegram.NewMessage) error {
	start := time.Now()
	updateLag := time.Since(time.Unix(int64(m.Date()), 0)).Milliseconds()

	msg, err := m.Reply("⏱️ Pinging...")
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)
	senders := m.Client.GetExportedSendersStatus()
	response := fmt.Sprintf(
		"<b>📊 System Performance Metrics</b>\n\n"+
			"⏱️ <b>Bot Latency:</b> <code>%d ms</code>\n"+
			"🕒 <b>Uptime:</b> <code>%s</code>\n"+
			"📩 <b>Update Lag:</b> <code>%d ms</code>\n"+
			"⚙️ <b>Go Routines:</b> <code>%d</code>\n"+
			"📨 <b>Senders:</b> <code>%d</code>\n",
		latency, uptime, updateLag, runtime.NumGoroutine(), senders,
	)

	_, err = msg.Edit(response)
	return err
}

// startHandler handles the /start command.
func startHandler(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	chatID := m.ChannelID()

	if m.IsPrivate() {
		go func(chatID int64) {
			ctx, cancel := db.Ctx()
			defer cancel()
			_ = db.Instance.AddUser(ctx, chatID)
		}(chatID)
	} else {
		go func(chatID int64) {
			ctx, cancel := db.Ctx()
			defer cancel()
			_ = db.Instance.AddChat(ctx, chatID)
		}(chatID)
	}

	response := fmt.Sprintf("🎶 <b>Welcome %s!</b>\n\nI am <b>%s</b> — your Spotify-style music bot for Telegram voice chat.\n\n<b>✨ Features:</b>\n• Auto custom thumbnail (Spotify vibe) for now playing\n• Fast queue and playback controls\n• Playlist support\n\n<b>🎧 Supported Platforms:</b>\nYouTube, Spotify, Apple Music, SoundCloud\n\nTap <b>Help</b> to see all commands.", m.Sender.FirstName, bot.FirstName)
		ReplyMarkup: core.AddMeMarkup(m.Client.Me().Username),
	})

	return err
}
