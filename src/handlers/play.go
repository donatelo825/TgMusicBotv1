/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */
TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ashokshau/tgmusic/config"
	"ashokshau/tgmusic/src/core"
	"ashokshau/tgmusic/src/core/cache"
	"ashokshau/tgmusic/src/core/db"
	"ashokshau/tgmusic/src/core/dl"
	"ashokshau/tgmusic/src/vc"

	"ashokshau/tgmusic/src/utils"

	"github.com/amarnathcjd/gogram/telegram"
)

// playHandler handles the /play command.
func playHandler(m *telegram.NewMessage) error {
	return handlePlay(m, false)
}

// vPlayHandler handles the /vplay command.
func vPlayHandler(m *telegram.NewMessage) error {
	return handlePlay(m, true)
}
@@ -175,53 +176,51 @@ func handleMedia(m *telegram.NewMessage, updater *telegram.NewMessage, dlMsg *te
	defer cancel()
	filePath, err := dlMsg.Download(&telegram.DownloadOptions{FileName: filepath.Join(config.Conf.DownloadsDir, fileName), Ctx: ctx})
	if err != nil {
		cache.ChatCache.RemoveCurrentSong(chatId) // Cleanup on failure
		_, err = updater.Edit(fmt.Sprintf("❌ Download failed: %s", err.Error()))
		return err
	}

	if dur == 0 {
		dur = utils.GetMediaDuration(filePath)
		saveCache.Duration = dur
	}

	saveCache.FilePath = filePath
	if err := vc.Calls.PlayMedia(chatId, saveCache.FilePath, saveCache.IsVideo, ""); err != nil {
		cache.ChatCache.RemoveCurrentSong(chatId)
		_, err = updater.Edit(err.Error())
		return err
	}

	nowPlaying := fmt.Sprintf(
		"🎵 <b>Now Playing:</b>\n\n<b>Track:</b> <a href='%s'>%s</a>\n<b>Duration:</b> %s\n<b>By:</b> %s",
		saveCache.URL, saveCache.Name, utils.SecToMin(saveCache.Duration), saveCache.User,
	)

	_, err = updater.Edit(nowPlaying, &telegram.SendOptions{
		ReplyMarkup: core.ControlButtons("play"),
	})
	err = sendNowPlayingCard(m, updater, nowPlaying, &saveCache)
	return err
}

// handleTextSearch handles a text search for a song.
func handleTextSearch(m *telegram.NewMessage, updater *telegram.NewMessage, wrapper *dl.DownloaderWrapper, chatId int64, isVideo bool, ctx context.Context) error {
	searchResult, err := wrapper.Search(ctx)
	if err != nil {
		_, err = updater.Edit(fmt.Sprintf("❌ Search failed: %s", err.Error()))
		return err
	}

	if searchResult.Results == nil || len(searchResult.Results) == 0 {
		_, err = updater.Edit("😕 No results found. Try a different query.")
		return err
	}

	song := searchResult.Results[0]
	if _track := cache.ChatCache.GetTrackIfExists(chatId, song.Id); _track != nil {
		_, err := updater.Edit("✅ Track already in queue or playing.")
		return err
	}

	return handleSingleTrack(m, updater, song, "", chatId, isVideo)
}

@@ -267,62 +266,94 @@ func handleSingleTrack(m *telegram.NewMessage, updater *telegram.NewMessage, son
	if saveCache.FilePath == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		dlResult, err := dl.DownloadSong(ctx, &saveCache, m.Client)
		if err != nil {
			cache.ChatCache.RemoveCurrentSong(chatId)
			_, err = updater.Edit(fmt.Sprintf("❌ Download failed: %s", err.Error()))
			return err
		}

		saveCache.FilePath = dlResult
	}

	if err := vc.Calls.PlayMedia(chatId, saveCache.FilePath, saveCache.IsVideo, ""); err != nil {
		cache.ChatCache.RemoveCurrentSong(chatId)
		_, err = updater.Edit(err.Error())
		return err
	}

	nowPlaying := fmt.Sprintf(
		"🎵 <b>Now Playing:</b>\n\n<b>Track:</b> <a href='%s'>%s</a>\n<b>Duration:</b> %s\n<b>By:</b> %s",
		saveCache.URL, saveCache.Name, utils.SecToMin(song.Duration), saveCache.User,
	)

	_, err := updater.Edit(nowPlaying, &telegram.SendOptions{
		ReplyMarkup: core.ControlButtons("play"),
	})
	err := sendNowPlayingCard(m, updater, nowPlaying, &saveCache)

	if err != nil {
		logger.Warn("Edit message failed: %v", err)
		return err
	}

	return nil
}

func sendNowPlayingCard(m *telegram.NewMessage, updater *telegram.NewMessage, caption string, track *utils.CachedTrack) error {
	thumbPath, err := utils.CreateSpotifyStyleCard(utils.ThumbnailCardOptions{
		Title:      track.Name,
		Subtitle:   track.Channel,
		CoverURL:   track.Thumbnail,
		OutputDir:  config.Conf.DownloadsDir,
		OutputName: fmt.Sprintf("now_playing_%d", time.Now().UnixNano()),
	})
	if err != nil {
		_, editErr := updater.Edit(caption, &telegram.SendOptions{ReplyMarkup: core.ControlButtons("play")})
		if editErr != nil {
			return editErr
		}
		return nil
	}
	defer os.Remove(thumbPath)

	_, sendErr := m.ReplyMedia(thumbPath, &telegram.MediaOptions{
		Caption:     caption,
		ParseMode:   "html",
		ReplyMarkup: core.ControlButtons("play"),
	})
	if sendErr != nil {
		_, editErr := updater.Edit(caption, &telegram.SendOptions{ReplyMarkup: core.ControlButtons("play")})
		if editErr != nil {
			return editErr
		}
		return nil
	}

	_, _ = updater.Delete()
	return nil
}

// handleMultipleTracks handles multiple tracks.
func handleMultipleTracks(m *telegram.NewMessage, updater *telegram.NewMessage, tracks []utils.MusicTrack, chatId int64, isVideo bool) error {
	if len(tracks) == 0 {
		_, err := updater.Edit("❌ No tracks found.")
		return err
	}

	queueHeader := "<b>📥 Added to Queue:</b>\n<blockquote collapsed='true'>\n"
	var tracksToAdd []*utils.CachedTrack
	var skippedTracks []string

	shouldPlayFirst := false
	var firstTrack *utils.CachedTrack

	for _, track := range tracks {
		if track.Duration > int(config.Conf.SongDurationLimit) {
			skippedTracks = append(skippedTracks, track.Title)
			continue
		}

		saveCache := &utils.CachedTrack{
			Name: track.Title, TrackID: track.Id, Duration: track.Duration,
			Thumbnail: track.Thumbnail, User: m.Sender.FirstName, Platform: track.Platform,
			IsVideo: isVideo, URL: track.Url, Channel: track.Channel, Views: track.Views,
		}

		}
		tracksToAdd = append(tracksToAdd, saveCache)
	}

	if len(tracksToAdd) == 0 {
		if len(skippedTracks) > 0 {
			_, err := updater.Edit(fmt.Sprintf("❌ All tracks were skipped (max duration %d min).", config.Conf.SongDurationLimit/60))
			return err
		}
		_, err := updater.Edit("❌ No valid tracks found.")
		return err
	}

	qLenAfter := cache.ChatCache.AddSongs(chatId, tracksToAdd)
	startLen := qLenAfter - len(tracksToAdd)

	if startLen == 0 {
		shouldPlayFirst = true
		firstTrack = tracksToAdd[0]
		firstTrack.Loop = 1
	}

	var sb strings.Builder
	sb.WriteString(queueHeader)

	totalDuration := 0
	for i, track := range tracksToAdd {
		currentQLen := startLen + i + 1
		fmt.Fprintf(&sb, "<b>%d.</b> %s\n└ Duration: %s\n",
			currentQLen, track.Name, utils.SecToMin(track.Duration))
		totalDuration += track.Duration
	}

	sb.WriteString("</blockquote>")
	queueSummary := fmt.Sprintf(
		"\n<b>📋 Queue Total:</b> %d\n<b>⏱ Duration:</b> %s\n<b>👤 By:</b> %s",
		qLenAfter, utils.SecToMin(totalDuration), m.Sender.FirstName,
	)

	sb.WriteString(queueSummary)
	if len(skippedTracks) > 0 {
		fmt.Fprintf(&sb, "\n\n<b>Skipped %d tracks</b> (exceeded duration limit).", len(skippedTracks))
	}

	fullMessage := sb.String()

	if len(fullMessage) > 4096 {
		fullMessage = queueSummary
	}

	if shouldPlayFirst && firstTrack != nil {
		_ = vc.Calls.PlayNext(chatId)
	}

	_, err := updater.Edit(fullMessage, &telegram.SendOptions{
		ReplyMarkup: core.ControlButtons("play"),
	})

	return err
}
