package mailer

import (
	"fmt"
	"net/smtp"
	"zeinai-backend/internal/config"
)

type Mailer struct {
	cfg config.SMTPConfig
}

func NewMailer(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) SendMail(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", m.cfg.From, to, subject, body)

	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	// In test/dev environment with mock settings, we might skip actual sending
	if m.cfg.User == "" || m.cfg.Password == "" {
		fmt.Printf("📧 MOCK EMAIL to %s: [%s] %s\n", to, subject, body)
		return nil
	}

	err := smtp.SendMail(addr, auth, m.cfg.User, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (m *Mailer) SendVerificationCode(to, code string) error {
	subject := "Код подтверждения - ZeinAI"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 10px;">
			<h2 style="color: #4a90e2; text-align: center;">Добро пожаловать в ZeinAI!</h2>
			<p>Ваш код подтверждения для регистрации:</p>
			<div style="background-color: #f5f7fa; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #2c3e50; border-radius: 5px; margin: 20px 0;">
				%s
			</div>
			<p style="color: #7f8c8d; font-size: 14px;">Если вы не запрашивали этот код, просто проигнорируйте это письмо.</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="text-align: center; color: #bdc3c7; font-size: 12px;">© 2025 ZeinAI EdTech Platform</p>
		</div>
	`, code)
	return m.SendMail(to, subject, body)
}

func (m *Mailer) SendPasswordResetCode(to, code string) error {
	subject := "Восстановление пароля - ZeinAI"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 10px;">
			<h2 style="color: #e67e22; text-align: center;">Восстановление доступа</h2>
			<p>Вы запросили сброс пароля. Используйте код ниже для установки нового пароля:</p>
			<div style="background-color: #f5f7fa; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 5px; color: #2c3e50; border-radius: 5px; margin: 20px 0;">
				%s
			</div>
			<p style="color: #7f8c8d; font-size: 14px;">Код действителен в течение 30 минут. Если вы не запрашивали сброс пароля, немедленно свяжитесь с поддержкой.</p>
			<hr style="border: 0; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="text-align: center; color: #bdc3c7; font-size: 12px;">© 2025 ZeinAI EdTech Platform</p>
		</div>
	`, code)
	return m.SendMail(to, subject, body)
}
