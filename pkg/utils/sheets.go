package utils

import (
	"context"
	"io"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	// URL Public Supabase kamu
	fileURL := "https://jmdmommnfxkcyauelsus.supabase.co/storage/v1/object/public/private_assets/service-account.json"

	// 1. Ambil file JSON secara langsung via HTTP
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 2. Baca seluruh isi body-nya (Raw Bytes)
	credsJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 3. Masukkan ke Google Sheets Service
	return sheets.NewService(ctx, option.WithCredentialsJSON(credsJSON))
}
