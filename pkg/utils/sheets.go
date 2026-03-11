package utils

import (
	"context"
	"io"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	fileURL := "https://jmdmommnfxkcyauelsus.supabase.co/storage/v1/object/public/private_assets/service-account.json"

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Baca murni tanpa diubah jadi string atau dimanipulasi
	credsJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, option.WithCredentialsJSON(credsJSON))
}
