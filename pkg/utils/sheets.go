package utils

import (
	"context"
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

	// SULAP TERAKHIR:
	// Mengubah teks literal "\n" menjadi karakter newline asli yang diminta Google
	content := string(body)
	content = strings.ReplaceAll(content, "\\n", "\n")

	return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(content)))
}
