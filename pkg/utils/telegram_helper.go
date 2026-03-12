package utils

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// EncodeTelegramPayload: email + otp -> base64 string
func EncodeTelegramPayload(email, otp string) string {
	str := fmt.Sprintf("%s|%s", email, otp)
	return base64.RawURLEncoding.EncodeToString([]byte(str))
}

// DecodeTelegramPayload: base64 string -> email, otp
func DecodeTelegramPayload(payload string) (string, string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid payload format")
	}
	return parts[0], parts[1], nil
}
