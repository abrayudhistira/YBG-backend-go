package utils

import (
	"context"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

//	func GetSheetsService(ctx context.Context, credentialsPath string) (*sheets.Service, error) {
//		srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
//		if err != nil {
//			return nil, err
//		}
//		return srv, nil
//	}
func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	// 1. Cek dulu di Environment Variable (Best Practice untuk Cloud)
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsJSON != "" {
		return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
	}

	// 2. Fallback ke file fisik (Hanya untuk kenyamanan saat coding di Laptop/Lokal)
	return sheets.NewService(ctx, option.WithCredentialsFile("config/service-account.json"))
}
