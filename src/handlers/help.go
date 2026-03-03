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
	"strings"

	"ashokshau/tgmusic/src/core"

	"github.com/amarnathcjd/gogram/telegram"
)

func getHelpCategories() map[string]struct {
	Title   string
	Content string
	Markup  *telegram.ReplyInlineMarkup
} {
	return map[string]struct {
		Title   string
		Content string
		Markup  *telegram.ReplyInlineMarkup
	}{
		"help_user": {
			Title:   "🎧 User Commands",
			Content: "<b>Playback:</b>\n• <code>/play [song]</code> — Play music\n\n<b>Utilities:</b>\n• <code>/start</code> — Start bot\n• <code>/privacy</code> — Privacy Policy\n• <code>/queue</code> — View queue",
			Markup:  core.BackHelpMenuKeyboard(),
		},
		"help_admin": {
			Title:   "⚙️ Admin Commands",
			Content: "<b>Controls:</b>\n• <code>/skip</code> — Skip track\n• <code>/pause</code> — Pause\n• <code>/resume</code> — Resume\n• <code>/seek [sec]</code> — Seek\n\n<b>Queue:</b>\n• <code>/remove [x]</code> — Remove track\n• <code>/loop [0-10]</code> — Loop queue\n\n<b>Access:</b>\n• <code>/auth [reply]</code> — Authorize user\n• <code>/unauth [reply]</code> — Unauthorize\n• <code>/authlist</code> — List authorized",
			Markup:  core.BackHelpMenuKeyboard(),
		},
		"help_devs": {
			Title:   "🛠 Developer Tools",
			Content: "<b>System:</b>\n• <code>/stats</code> — Usage stats\n\n<b>Maintenance:</b>\n• <code>/av</code> — Active voice chats",
			Markup:  core.BackHelpMenuKeyboard(),
		},
		"help_owner": {
			Title:   "🔐 Owner Commands",
			Content: "<b>Settings:</b>\n• <code>/settings</code> — Chat settings",
			Markup:  core.BackHelpMenuKeyboard(),
		},
		"help_playlist": {
			Title:   "🎵 Playlist Commands",
			Content: "<b>Playlist Management:</b>\n• <code>/createplaylist [name]</code> — Create playlist\n• <code>/deleteplaylist [id]</code> — Delete playlist\n• <code>/addtoplaylist [id] [url]</code> — Add song\n• <code>/removefromplaylist [id] [url]</code> — Remove song\n• <code>/playlistinfo [id]</code> — Playlist info\n• <code>/myplaylists</code> — My playlists",
			Markup:  core.BackHelpMenuKeyboard(),
		},
	}
}

// helpCallbackHandler handles callbacks from the help keyboard.
// It takes a telegram.CallbackQuery object as input.
// It returns an error if any.
func helpCallbackHandler(cb *telegram.CallbackQuery) error {
	data := cb.DataString()

	helpCategories := getHelpCategories()
	if strings.Contains(data, "help_all") {
		_, _ = cb.Answer("📚 Opening Help Menu...", &telegram.CallbackOptions{Alert: false}
			response := fmt.Sprintf("🎶 <b>Welcome %s!</b>\n\nI am <b>%s</b> — your Spotify-style music bot for Telegram voice chat.\n\n<b>✨ Features:</b>\n• Auto custom thumbnail (Spotify vibe) for now playing\n• Fast queue and playback controls\n• Playlist support\n\n<b>🎧 Supported Platforms:</b>\nYouTube, Spotify, Apple Music, SoundCloud\n\nTap <b>Help</b> to see all commands.", cb.Sender.FirstName, cb.Client.Me().FirstName)
		_, _ = cb.Edit(response, &telegram.SendOptions{ReplyMarkup: core.HelpMenuKeyboard()})
		return nil
	}

	if strings.Contains(data, "help_back") {
		_, _ = cb.Answer("🏠 Returning to home...", &telegram.CallbackOptions{Alert: false})
		response := fmt.Sprintf("🎶 <b>Welcome %s!</b>\n\nI am <b>%s</b> — your Spotify-style music bot for Telegram voice chat.\n\n<b>✨ Features:</b>\n• Auto custom thumbnail (Spotify vibe) for now playing\n• Fast queue and playback controls\n• Playlist support\n\n<b>🎧 Supported Platforms:</b>\nYouTube, Spotify, Apple Music, SoundCloud\n\nTap <b>Help</b> to see all commands.", cb.Sender.FirstName, cb.Client.Me().FirstName)
		_, _ = cb.Edit(response, &telegram.SendOptions{ReplyMarkup: core.AddMeMarkup(cb.Client.Me().Username)})
		return nil
	}

	if category, ok := helpCategories[data]; ok {
		_, _ = cb.Answer(fmt.Sprintf("📖 %s", category.Title), &telegram.CallbackOptions{Alert: false})
		text := fmt.Sprintf("<b>%s</b>\n\n%s\n\n🔙 <i>Use buttons below to go back.</i>", category.Title, category.Content)
		_, _ = cb.Edit(text, &telegram.SendOptions{ReplyMarkup: category.Markup})
		return nil
	}

	_, _ = cb.Answer("⚠️ Unknown command category.", &telegram.CallbackOptions{Alert: false})
	return nil
}

// privacyHandler handles the /privacy command.
// It takes a telegram.NewMessage object as input.
// It returns an error if any.
func privacyHandler(m *telegram.NewMessage) error {
	botName := m.Client.Me().FirstName

	text := fmt.Sprintf("<b>Privacy Policy for %s</b>\n\n<b>1. Data Storage:</b>\nWe do not store personal data on your device. We do not track your browsing activity.\n\n<b>2. Collection:</b>\nWe only collect your Telegram <b>User ID</b> and <b>Chat ID</b> to provide music services. No names, phone numbers, or locations are stored.\n\n<b>3. Usage:</b>\nData is used strictly for bot functionality. No marketing or commercial use.\n\n<b>4. Sharing:</b>\nWe do not share data with third parties. No data is sold or traded.\n\n<b>5. Security:</b>\nWe use standard encryption to protect data. However, no online service is 100%% secure.\n\n<b>6. Cookies:</b>\n%s does not use cookies or tracking technologies.\n\n<b>7. Third Parties:</b>\nWe do not integrate with third-party data collectors, other than Telegram itself.\n\n<b>8. Your Rights:</b>\nYou can request data deletion or block the bot to revoke access.\n\n<b>9. Updates:</b>\nPolicy changes will be announced in the bot.\n\n<b>10. Contact:</b>\nQuestions? Contact our <a href=\"https://t.me/GuardxSupport\">Support Group</a>.\n\n──────────────────\n<b>Note:</b> This policy ensures a safe and respectful experience with %s.", botName, botName, botName)

	_, err := m.Reply(text, &telegram.SendOptions{LinkPreview: false})
	return err
}
