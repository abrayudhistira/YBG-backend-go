package utils

import (
	"context"
	"os"
	"strings" // Tambahkan ini

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")

	if credsJSON != "" {
		// TRIK SAKTI: Memastikan karakter \n dibaca sebagai baris baru beneran oleh Google
		credsJSON = strings.ReplaceAll(credsJSON, "\\n", "\n")

		return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
	}

	// Fallback lokal
	return sheets.NewService(ctx, option.WithCredentialsFile("config/service-account.json"))
}
