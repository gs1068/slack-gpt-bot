package spreadsheet

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gs1068/slack-gpt-bot/domain/model"
	"github.com/gs1068/slack-gpt-bot/domain/repository"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	dataRange = "Activity!A:E"
)

var SpreadsheetID string

type SpreadsheetRepository struct {
	ssClient *sheets.Service
}

func NewSpreadsheetRepository(ssClient *sheets.Service) repository.SpreadsheetRepository {
	return &SpreadsheetRepository{
		ssClient: ssClient,
	}
}

func SpreadSheetClient() (*sheets.Service, error) {
	ctx := context.Background()
	credentialsFilePath := "./credential.json"
	b, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed os.ReadFile: %w", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed google.JWTConfigFromJSON: %w", err)
	}

	client := config.Client(ctx)
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

func (r *SpreadsheetRepository) GetSpreadsheetDataBySlackID(ctx context.Context, userID string) (*model.SpreadsheetData, error) {
	values, err := r.readSpreadsheet(ctx, dataRange)
	if err != nil {
		return nil, fmt.Errorf("failed r.readSpreadsheet: %w", err)
	}

	for _, row := range values {
		if len(row) < 5 {
			continue
		}

		if row[0].(string) == userID {
			totalUsage, _ := strconv.Atoi(row[1].(string))
			tokensUsage, _ := strconv.Atoi(row[3].(string))
			dailyTokensUsage, _ := strconv.Atoi(row[4].(string))
			return &model.SpreadsheetData{
				UserID:           userID,
				TotalUsage:       totalUsage,
				LastUsedAt:       row[2].(string),
				TokensUsage:      tokensUsage,
				DailyTokensUsage: dailyTokensUsage,
				TotalTokensUsage: tokensUsage,
			}, nil
		}
	}

	return nil, nil
}

func (r *SpreadsheetRepository) UpdateSpreadsheet(ctx context.Context, update model.SpreadsheetData) error {
	values, err := r.readSpreadsheet(ctx, dataRange)
	if err != nil {
		return fmt.Errorf("failed r.readSpreadsheet: %w", err)
	}

	userData := r.mapSpreadsheetData(values)
	now := time.Now().Format(time.RFC3339)
	userData[update.UserID] = r.convertActivityData(update, now)

	updatedValues := r.mapToSortedSlice(userData)

	err = r.writeSpreadsheet(ctx, dataRange, updatedValues)
	if err != nil {
		return fmt.Errorf("failed r.writeSpreadsheet: %w", err)
	}

	return nil
}

func (r *SpreadsheetRepository) mapSpreadsheetData(values [][]interface{}) map[string][]interface{} {
	userData := make(map[string][]interface{}, len(values))
	for _, row := range values {
		if len(row) < 5 {
			continue
		}

		userID, ok := row[0].(string)
		if !ok {
			continue
		}
		userData[userID] = row
	}
	return userData
}

func (r *SpreadsheetRepository) convertActivityData(update model.SpreadsheetData, now string) []interface{} {
	return []interface{}{
		update.UserID,
		fmt.Sprintf("%d", update.TotalUsage),
		now,
		fmt.Sprintf("%d", update.TotalTokensUsage),
		fmt.Sprintf("%d", update.DailyTokensUsage),
	}
}

func (r *SpreadsheetRepository) mapToSortedSlice(userData map[string][]interface{}) [][]interface{} {
	updatedValues := make([][]interface{}, 0, len(userData))
	for _, row := range userData {
		updatedValues = append(updatedValues, row)
	}
	sort.Slice(updatedValues, func(i, j int) bool {
		return updatedValues[i][0].(string) < updatedValues[j][0].(string)
	})
	return updatedValues
}

func (r *SpreadsheetRepository) readSpreadsheet(ctx context.Context, readRange string) ([][]interface{}, error) {
	resp, err := r.ssClient.Spreadsheets.Values.Get(SpreadsheetID, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to r.ssClient.Spreadsheets.Values.Get: %w", err)
	}
	return resp.Values, nil
}

func (r *SpreadsheetRepository) writeSpreadsheet(ctx context.Context, writeRange string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}
	_, err := r.ssClient.Spreadsheets.Values.Update(SpreadsheetID, writeRange, valueRange).
		ValueInputOption("RAW").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("failed r.ssClient.Spreadsheets.Values.Update: %w", err)
	}
	return nil
}
