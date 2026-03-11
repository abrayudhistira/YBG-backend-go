package utils

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func GetSheetsService(ctx context.Context) (*sheets.Service, error) {
	// Hardcode JSON string
	credsJSON := `{
  "type": "service_account",
  "project_id": "digiwaste-459318",
  "private_key_id": "4273bc8f45849040708af556010160bcf68e9f54",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDV9JrukfWhsjp+\nMc7teDNpLPPIVG4CuNBukBRzfFWaCKz7BU4lIyzYaYG1PQDXrlYuCz3vq4f5TbA5\n7poN2d6khtpW2hquaTHKF7NkW2FFju8iwPxIgSphVFnvZqbyTJt2L/NRNjWatqEZ\nfErqDsa0QRtPkLhFnseMhu3fg6gUVx4e7vZL/Qs5MOpVLbDiDkU85Mt66z6ci7rM\nSIWKinhRTbnvtYiC+0GBZi5/VzHGAdYUKbAaHsT2TuGHWwQrUXviC6GFFlc3dl0N\nlKbRpqOU+4KLIOdzc72CGNe/4Y+WhkcGhUHkyxpadozOWT3a97ro9ejIROPNwVv0\ncTSrGQbNAgMBAAECggEAVcfFrDvZ2vPtrrXCjIQCPLtYnCt5ld7KNmHOyUSCv4iV\n7eiBHbOeIcAfUG4+XbrYc4pvUR2ZHROQQZHPsxj0QkuM04CLbPzhCPD6rBRVCgHW\nD72HCHy85JvgmPKzoXakZ7yu1ZMh578sFN832+KDuTZXQE26C7OutsFMMq6C33AY\nhHJTbSvrb3EpZ0mk+vR3e6CZJFDCXU2zfP1eAzj9ISC/3F3TuYDg2B9A15hrsb+g\n8MlvZbnfKA4rbtNcxTTMIVtk6H+TA8nWd1XwC+zqQzw3OgDsYsUMK1Rs8g0UaNW9\n5aefCBabUlvR0m9LuPs12TGkyIZH2GtQOkIuZHeUUwKBgQDtjPLrAa7H7XwkyYNF\nRTFLyWoVlrMnGYmZGASefyJYCbHocR5Jr8yJr7rnn+k+kfCxzjhlQHg8S00bBEE1\nKsZrcQqJ6DTFYuL/7Xn8V+MUCXOhAYSS37Lo1J2gG1lBpnlPTWSLDKImzqLzKo6M\ndDKoeXfdxm3XdJsZIvMKqJoVRwKBgQDmkogcCJJeo/xauEhtqaQ7dpskXgQqS6TA\nsSxb1j6X7mxWGqzkriTPnHjGWH9N38reewwqmrzmXcgJ/dlUPF+x+jumQokFYfid\nUx04PwO2DC5sAvAP5IDQxt2VgtTZnOwws2OXuTSTVdH4qPWQrVpfxpWYwpahHGLe\nyX/pVvZdSwKBgB2BQDrIPrk+WgkHrnJQIctT/QUpbp8QoPKO9SPqjo14xswkIKru\Vu1TElfqmMHYxpiPEJoi48w5Xh5Y7PB5m6OEqtZuLP/HRIKdMGWTVPUMJ3x7/8du\nWX5pyho0y2VIFBExf6d1rj47tCmXw5TWaeRbEfRNzR7RsOHyYVWVk23JAoGADws4\nfjvA8RPZ/0FO2HjdElQmwzSvKONOmJP2xPcxllAkGWocJb+G/1TCPI7Bn58eaW21\n2YHHGXC9AInjiC94PvCIu8xTjFpcEke9/FGAOHyK+tkmOKM8FGMlSgADSz+F2Zea\nw+d9mq9ax9KeUxY8c0tNr23izhhACzEye1MFOAkCgYAX2PEyp3cToNYEU66Fr8bs\ndWPwLoT59PBDL+ajc0cQDqbABNicjWXmXxAhKdsDWoGh7r19asjqszZQ6/UdPA8d\nY5ps/bXl1KBX2vOBuzeLgi9DVgWqEjNb2BDwwZIVsDc2zEVfssdqO7oG0WkMLHgK\ntjGl+gMyoPMBadlj4NP//g==\n-----END PRIVATE KEY-----\n",
  "client_email": "spreadsheet-worker@digiwaste-459318.iam.gserviceaccount.com",
  "client_id": "116948120066161969897",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/spreadsheet-worker%40digiwaste-459318.iam.gserviceaccount.com",
  "universe_domain": "googleapis.com"
}`

	return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credsJSON)))
}
