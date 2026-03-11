package utils

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context, credentialsPath string) (*sheets.Service, error) {
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, err
	}
	return srv, nil
}
