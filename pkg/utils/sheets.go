package utils

import (
	"context"
	"os"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")

	if credsJSON != "" {
		// 1. Pastikan tidak ada karakter "enter" fisik yang merusak JSON
		// Kita bersihkan spasi/enter di awal dan akhir string
		credsJSON = strings.TrimSpace(credsJSON)

		// 2. Jika kamu paste di Vercel dalam bentuk "One Line" tapi private_key-nya
		// mengandung \n, Google akan membacanya dengan benar lewat option ini.
		return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
	}

	// Fallback untuk development lokal (pastikan path-nya benar)
	return sheets.NewService(ctx, option.WithCredentialsFile("config/service-account.json"))
}
