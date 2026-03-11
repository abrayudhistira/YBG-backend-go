package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 1. Masukkan ke Map dulu supaya format JSON-nya nggak rusak
	var creds map[string]interface{}
	if err := json.Unmarshal(body, &creds); err != nil {
		return nil, err
	}

	// 2. Ambil private_key dan benerin \n nya di sini
	if pk, ok := creds["private_key"].(string); ok {
		creds["private_key"] = strings.ReplaceAll(pk, "\\n", "\n")
	}

	// 3. Encode balik jadi JSON yang valid
	finalJSON, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, option.WithCredentialsJSON(finalJSON))
}
