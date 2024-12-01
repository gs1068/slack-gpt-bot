package repository

import (
	"context"

	"github.com/gs1068/slack-gpt-bot/domain/model"
)

type SpreadsheetRepository interface {
	GetSpreadsheetDataBySlackID(ctx context.Context, userID string) (*model.SpreadsheetData, error)
	UpdateSpreadsheet(ctx context.Context, update model.SpreadsheetData) error
}
