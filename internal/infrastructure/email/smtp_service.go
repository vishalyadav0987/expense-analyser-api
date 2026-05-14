package email

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
)

type SMTPEmailService struct {
	from     string
	password string
	host     string
	port     string
}

func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		from:     os.Getenv("EMAIL_ID"),
		password: os.Getenv("PASSWORD"),
		host:     os.Getenv("HOST"),
		port:     os.Getenv("EMAIL_SERVICE_PORT"),
	}
}

func (s *SMTPEmailService) SendOTP(ctx context.Context, email, otp string) error {
	auth := smtp.PlainAuth("", os.Getenv("BREVO_EMAIL_API"), s.password, s.host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: OTP Verification\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>OTP Verification</title>
	</head>
	<body style="margin:0; padding:0; background-color:#f4f6f8; font-family:Arial, sans-serif;">

		<div style="max-width:500px; margin:40px auto; background:#ffffff; border-radius:10px; overflow:hidden; box-shadow:0 4px 12px rgba(0,0,0,0.1);">

			<div style="background:#4f46e5; padding:20px; text-align:center; color:white;">
				<h2 style="margin:0;">Your Verification Code</h2>
			</div>

			<div style="padding:30px; text-align:center;">
				<p style="font-size:16px; color:#333;">
					Use the OTP below to complete your verification
				</p>

				<div style="margin:25px 0; font-size:28px; letter-spacing:6px; font-weight:bold; color:#111;">
					%s
				</div>

				<p style="font-size:13px; color:#777;">
					This OTP is valid for <b>5 minutes</b>. Do not share it with anyone.
				</p>

				<div style="margin-top:30px; font-size:12px; color:#999;">
					If you didn’t request this, you can ignore this email.
				</div>
			</div>

			<div style="background:#f0f0f0; padding:15px; text-align:center; font-size:12px; color:#666;">
				© 2026 Your Company. All rights reserved.
			</div>

		</div>

	</body>
	</html>
	`,
		s.from, email, otp,
	))

	err := smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.from,
		[]string{email},
		msg,
	)

	if err != nil {
		fmt.Println("EMAIL ERROR:", err)
	}
	return err
}
