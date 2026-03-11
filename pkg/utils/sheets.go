package utils

import (
	"context"
	"encoding/base64"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	// Ambil dari ENV Vercel
	data := os.Getenv("GOOGLE_CREDENTIALS_JSON")

	credsJSON, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, option.WithCredentialsJSON(credsJSON))
}
