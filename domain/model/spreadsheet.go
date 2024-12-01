package model

import (
	"errors"
	"time"
)

type SpreadsheetID string

const (
	DailyTokenLimit = 20000
)

type SpreadsheetData struct {
	UserID           string
	TotalUsage       int
	LastUsedAt       string
	TokensUsage      int
	DailyTokensUsage int
	TotalTokensUsage int
}

func NewSpreadsheet(
	userID string,
	totalUsage int,
	lastUsedAt string,
	tokensUsage int,
	dailyTokensUsage int,
	totalTokensUsage int,
) *SpreadsheetData {
	return &SpreadsheetData{
		UserID:           userID,
		TotalUsage:       totalUsage,
		LastUsedAt:       lastUsedAt,
		TokensUsage:      tokensUsage,
		DailyTokensUsage: dailyTokensUsage,
		TotalTokensUsage: totalTokensUsage,
	}
}

func (s *SpreadsheetData) CanUseDailyTokens() error {
	if s.DailyTokensUsage > DailyTokenLimit {
		return errors.New("daily token limit exceeded")
	}
	return nil
}

func (s *SpreadsheetData) AddTokenUsage(tokens int) {
	s.TotalUsage++
	s.DailyTokensUsage += tokens
	s.TotalTokensUsage += tokens
}

func (s *SpreadsheetData) ResetDailyUsageIfNeeded() {
	now := time.Now().In(time.FixedZone("Asia/Tokyo", 9*60*60))
	nowDate := now.Format("2006-01-02")

	if s.LastUsedAt != "" {
		lastUsedTime, _ := time.Parse(time.RFC3339, s.LastUsedAt)
		lastUsedDate := lastUsedTime.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format("2006-01-02")
		if lastUsedDate != nowDate {
			s.DailyTokensUsage = 0
		}
	}

	s.LastUsedAt = now.Format(time.RFC3339)
}
