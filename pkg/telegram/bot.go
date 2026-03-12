package telegram

import (
	"fmt"
	"ybg-backend-go/core/usecase"
	"ybg-backend-go/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotService struct {
	bot *tgbotapi.BotAPI
	uc  usecase.UserUsecase
}

func NewBotService(token string, uc usecase.UserUsecase) *BotService {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		// Kita jangan Panic di sini agar server Vercel tidak mati total
		// Cukup print error saja
		fmt.Printf("Telegram Bot Error: %v\n", err)
		return &BotService{bot: nil, uc: uc}
	}
	return &BotService{bot: bot, uc: uc}
}

// GANTI ATAU TAMBAHKAN METHOD INI
func (s *BotService) HandleUpdate(update tgbotapi.Update) {
	if s.bot == nil || update.Message == nil || !update.Message.IsCommand() {
		return
	}

	if update.Message.Command() == "start" {
		payload := update.Message.CommandArguments()
		if payload == "" {
			s.reply(update.Message.Chat.ID, "Halo! Silakan gunakan link dari aplikasi iCoass untuk reset password kamu.")
			return
		}

		// 1. Decode Payload (Email | OTP)
		email, otp, err := utils.DecodeTelegramPayload(payload)
		if err != nil {
			s.reply(update.Message.Chat.ID, "Maaf, link verifikasi tidak valid atau sudah kadaluarsa.")
			return
		}

		// 2. Validasi & Generate Reset Token lewat Usecase
		resetToken, err := s.uc.ValidateOTPAndGenerateResetToken(email, otp)
		if err != nil {
			s.reply(update.Message.Chat.ID, "Gagal verifikasi: "+err.Error())
			return
		}

		// 3. Kirim link Reset ke User
		// Sesuaikan URL ini dengan domain frontend/api kamu
		link := "https://ybg-backend-go.vercel.app/reset-password?token=" + resetToken
		msg := fmt.Sprintf("✅ Verifikasi Berhasil!\n\nKlik link di bawah ini untuk mengganti password kamu (berlaku 15 menit):\n\n%s", link)

		s.reply(update.Message.Chat.ID, msg)
	}
}

func (s *BotService) reply(chatID int64, text string) {
	if s.bot == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	s.bot.Send(msg)
}
