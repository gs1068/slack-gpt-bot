package model

import (
	"testing"
	"time"
)

func TestCanUseDailyTokens(t *testing.T) {
	tests := []struct {
		name        string
		dailyTokens int
		expectedErr bool
	}{
		{
			name:        "within limit",
			dailyTokens: 15000,
			expectedErr: false,
		},
		{
			name:        "exceeds limit",
			dailyTokens: 20001,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spreadsheet := &SpreadsheetData{
				DailyTokensUsage: tt.dailyTokens,
			}
			err := spreadsheet.CanUseDailyTokens()
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestResetDailyUsageIfNeeded(t *testing.T) {
	tests := []struct {
		name             string
		lastUsedAt       string
		expectedDailyUse int
	}{
		{
			name:             "same day usage",
			lastUsedAt:       time.Now().In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339),
			expectedDailyUse: 100,
		},
		{
			name:             "different day usage",
			lastUsedAt:       time.Now().AddDate(0, 0, -1).In(time.FixedZone("Asia/Tokyo", 9*60*60)).Format(time.RFC3339),
			expectedDailyUse: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spreadsheet := &SpreadsheetData{
				LastUsedAt:       tt.lastUsedAt,
				DailyTokensUsage: 100,
			}
			spreadsheet.ResetDailyUsageIfNeeded()
			if spreadsheet.DailyTokensUsage != tt.expectedDailyUse {
				t.Errorf("expected daily usage: %v, got: %v", tt.expectedDailyUse, spreadsheet.DailyTokensUsage)
			}
		})
	}
}
