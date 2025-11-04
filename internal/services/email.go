package services

import (
	"fmt"

	"github.com/Qovix/Qovix-go/internal/config"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	log    *logrus.Logger
	config *config.EmailConfig
}

func NewEmailService(emailConfig *config.EmailConfig) *EmailService {
	return &EmailService{
		log:    logger.GetLogger(),
		config: emailConfig,
	}
}

func (s *EmailService) SendVerificationEmail(email, code string) error {
	subject := "Verify Your Email - Qovix"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07); overflow: hidden;">
                    <!-- Header with gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #bc3a08 0%%, #d94e1a 100%%; padding: 40px 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 32px; font-weight: 700; letter-spacing: -0.5px;">Qovix</h1>
                            <div style="width: 60px; height: 4px; background-color: rgba(255, 255, 255, 0.5); margin: 20px auto 0; border-radius: 2px;"></div>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 50px 40px;">
                            <h2 style="margin: 0 0 10px; color: #1a1a1a; font-size: 26px; font-weight: 600;">Verify Your Email</h2>
                            <p style="margin: 0 0 30px; color: #666666; font-size: 16px; line-height: 1.6;">
                                Thank you for signing up! To complete your registration and start building amazing database schemas, please verify your email address using the code below:
                            </p>
                            
                            <!-- Verification Code Box -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
                                <tr>
                                    <td align="center" style="padding: 30px 0;">
                                        <div style="background: linear-gradient(135deg, #fff5f2 0%%, #ffe8e0 100%%); border: 2px dashed #bc3a08; border-radius: 12px; padding: 25px; display: inline-block;">
                                            <div style="color: #bc3a08; font-size: 36px; font-weight: 800; letter-spacing: 12px; font-family: 'Courier New', monospace;">%s</div>
                                        </div>
                                    </td>
                                </tr>
                            </table>
                            
                            <!-- Info boxes -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-top: 30px;">
                                <tr>
                                    <td style="background-color: #fff9f7; border-left: 4px solid #bc3a08; padding: 16px 20px; border-radius: 8px;">
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                            ⏱️ <strong style="color: #bc3a08;">Expires in 15 minutes</strong> – Please complete verification soon.
                                        </p>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="margin: 30px 0 0; color: #999999; font-size: 14px; line-height: 1.6;">
                                If you didn't create an account with Qovix, you can safely ignore this email.
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #fafafa; padding: 30px 40px; text-align: center; border-top: 1px solid #eeeeee;">
                            <p style="margin: 0 0 8px; color: #999999; font-size: 13px;">
                                © 2024 Qovix. All rights reserved.
                            </p>
                            <p style="margin: 0; color: #cccccc; font-size: 12px;">
                                Design better databases, faster.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
    `, code)

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) SendPasswordResetEmail(email, code string) error {
	subject := "Password Reset - Qovix"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07); overflow: hidden;">
                    <!-- Header with gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #bc3a08 0%%, #d94e1a 100%%; padding: 40px 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 32px; font-weight: 700; letter-spacing: -0.5px;">Schema Builder</h1>
                            <div style="width: 60px; height: 4px; background-color: rgba(255, 255, 255, 0.5); margin: 20px auto 0; border-radius: 2px;"></div>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 50px 40px;">
                            <div style="text-align: center; margin-bottom: 30px;">
                                <div style="display: inline-block; background-color: #fff5f2; border-radius: 50%%; padding: 20px; margin-bottom: 20px;">
                                    <span style="font-size: 40px;">🔐</span>
                                </div>
                            </div>
                            
                            <h2 style="margin: 0 0 10px; color: #1a1a1a; font-size: 26px; font-weight: 600; text-align: center;">Reset Your Password</h2>
                            <p style="margin: 0 0 30px; color: #666666; font-size: 16px; line-height: 1.6; text-align: center;">
                                We received a request to reset your password. Use the code below to create a new password:
                            </p>
                            
                            <!-- Reset Code Box -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
                                <tr>
                                    <td align="center" style="padding: 30px 0;">
                                        <div style="background: linear-gradient(135deg, #fff5f2 0%%, #ffe8e0 100%%); border: 2px dashed #bc3a08; border-radius: 12px; padding: 25px; display: inline-block;">
                                            <div style="color: #bc3a08; font-size: 36px; font-weight: 800; letter-spacing: 12px; font-family: 'Courier New', monospace;">%s</div>
                                        </div>
                                    </td>
                                </tr>
                            </table>
                            
                            <!-- Warning box -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-top: 30px;">
                                <tr>
                                    <td style="background-color: #fff9f7; border-left: 4px solid #bc3a08; padding: 16px 20px; border-radius: 8px;">
                                        <p style="margin: 0 0 10px; color: #666666; font-size: 14px; line-height: 1.5;">
                                            ⏱️ <strong style="color: #bc3a08;">Expires in 15 minutes</strong> – Use this code soon.
                                        </p>
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                            🔒 <strong style="color: #bc3a08;">Security tip:</strong> Never share this code with anyone.
                                        </p>
                                    </td>
                                </tr>
                            </table>
                            
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-top: 20px;">
                                <tr>
                                    <td style="background-color: #f0f0f0; padding: 16px 20px; border-radius: 8px; text-align: center;">
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">
                                            Didn't request a password reset? You can safely ignore this email.
                                        </p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #fafafa; padding: 30px 40px; text-align: center; border-top: 1px solid #eeeeee;">
                            <p style="margin: 0 0 8px; color: #999999; font-size: 13px;">
                                © 2024 Schema Builder. All rights reserved.
                            </p>
                            <p style="margin: 0; color: #cccccc; font-size: 12px;">
                                Design better databases, faster.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
    `, code)

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) SendWelcomeEmail(email, firstName string) error {
	subject := "Welcome to Qovix!"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07); overflow: hidden;">
                    <!-- Header with gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #bc3a08 0%%, #d94e1a 100%%; padding: 40px 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 32px; font-weight: 700; letter-spacing: -0.5px;">Schema Builder</h1>
                            <div style="width: 60px; height: 4px; background-color: rgba(255, 255, 255, 0.5); margin: 20px auto 0; border-radius: 2px;"></div>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 50px 40px;">
                            <div style="text-align: center; margin-bottom: 30px;">
                                <div style="display: inline-block; background: linear-gradient(135deg, #fff5f2 0%%, #ffe8e0 100%%); border-radius: 50%%; padding: 25px; margin-bottom: 20px;">
                                    <span style="font-size: 48px;">🎉</span>
                                </div>
                            </div>
                            
                            <h2 style="margin: 0 0 10px; color: #1a1a1a; font-size: 28px; font-weight: 700; text-align: center;">Welcome to Qovix!</h2>
                            <p style="margin: 0 0 30px; color: #bc3a08; font-size: 18px; font-weight: 600; text-align: center;">
                                Hi %s! 👋
                            </p>
                            
                            <p style="margin: 0 0 25px; color: #666666; font-size: 16px; line-height: 1.7; text-align: center;">
                                Your email has been successfully verified and your account is now <strong style="color: #bc3a08;">active</strong>! You're all set to start building amazing database schemas.
                            </p>
                            
                            <!-- Feature highlights -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 35px 0;">
                                <tr>
                                    <td style="padding: 20px; background-color: #fff9f7; border-radius: 10px; margin-bottom: 12px;">
                                        <p style="margin: 0 0 8px; color: #bc3a08; font-size: 20px;">✨</p>
                                        <p style="margin: 0 0 5px; color: #1a1a1a; font-size: 15px; font-weight: 600;">Design Database Schemas</p>
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">Create visual database models with our intuitive drag-and-drop interface.</p>
                                    </td>
                                </tr>
                            </table>
                            
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 0 0 12px;">
                                <tr>
                                    <td style="padding: 20px; background-color: #fff9f7; border-radius: 10px;">
                                        <p style="margin: 0 0 8px; color: #bc3a08; font-size: 20px;">🤝</p>
                                        <p style="margin: 0 0 5px; color: #1a1a1a; font-size: 15px; font-weight: 600;">Collaborate with Teams</p>
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">Work together in real-time with your team members and stakeholders.</p>
                                    </td>
                                </tr>
                            </table>
                            
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 0 0 35px;">
                                <tr>
                                    <td style="padding: 20px; background-color: #fff9f7; border-radius: 10px;">
                                        <p style="margin: 0 0 8px; color: #bc3a08; font-size: 20px;">📤</p>
                                        <p style="margin: 0 0 5px; color: #1a1a1a; font-size: 15px; font-weight: 600;">Export Your Work</p>
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.5;">Generate SQL, documentation, and diagrams in multiple formats.</p>
                                    </td>
                                </tr>
                            </table>
                            
                            <!-- CTA Button -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
                                <tr>
                                    <td align="center" style="padding: 20px 0;">
                                        <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #bc3a08 0%%, #d94e1a 100%%); color: #ffffff; text-decoration: none; padding: 16px 48px; border-radius: 10px; font-size: 16px; font-weight: 600; box-shadow: 0 4px 12px rgba(188, 58, 8, 0.3);">
                                            Get Started Now →
                                        </a>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="margin: 30px 0 0; color: #999999; font-size: 14px; line-height: 1.6; text-align: center;">
                                Need help getting started? Our support team is here for you!
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #fafafa; padding: 30px 40px; text-align: center; border-top: 1px solid #eeeeee;">
                            <p style="margin: 0 0 8px; color: #999999; font-size: 13px;">
                                © 2024 Schema Builder. All rights reserved.
                            </p>
                            <p style="margin: 0; color: #cccccc; font-size: 12px;">
                                Design better databases, faster.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
    `, firstName, "http://localhost:5173/dashboard")

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) SendAccountLinkingNotification(email, firstName string) error {
	subject := "Google Account Linked - Qovix"
	body := fmt.Sprintf(`
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #f5f5f5; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="max-width: 600px; background-color: #ffffff; border-radius: 16px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07); overflow: hidden;">
                    <!-- Header with gradient -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #bc3a08 0%%, #d94e1a 100%%; padding: 40px 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 32px; font-weight: 700; letter-spacing: -0.5px;">Schema Builder</h1>
                            <div style="width: 60px; height: 4px; background-color: rgba(255, 255, 255, 0.5); margin: 20px auto 0; border-radius: 2px;"></div>
                        </td>
                    </tr>
                    
                    <!-- Content -->
                    <tr>
                        <td style="padding: 50px 40px;">
                            <div style="text-align: center; margin-bottom: 30px;">
                                <div style="display: inline-block; background: linear-gradient(135deg, #fff5f2 0%%, #ffe8e0 100%%); border-radius: 50%%; padding: 25px; margin-bottom: 20px;">
                                    <span style="font-size: 48px;">🔗</span>
                                </div>
                            </div>
                            
                            <h2 style="margin: 0 0 10px; color: #1a1a1a; font-size: 26px; font-weight: 600; text-align: center;">Account Successfully Linked</h2>
                            <p style="margin: 0 0 30px; color: #666666; font-size: 16px; line-height: 1.6;">
                                Hi %s,
                            </p>
                            
                            <p style="margin: 0 0 25px; color: #666666; font-size: 16px; line-height: 1.7;">
                                Great news! Your <strong style="color: #bc3a08;">Google account</strong> has been successfully linked to your Qovix account.
                            </p>
                            
                            <!-- Benefits box -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin: 30px 0;">
                                <tr>
                                    <td style="background: linear-gradient(135deg, #fff5f2 0%%, #ffe8e0 100%%); border-radius: 12px; padding: 25px;">
                                        <p style="margin: 0 0 15px; color: #bc3a08; font-size: 16px; font-weight: 700;">
                                            ✓ What this means for you:
                                        </p>
                                        <p style="margin: 0 0 10px; color: #666666; font-size: 15px; line-height: 1.6;">
                                            • Sign in faster with your Google account
                                        </p>
                                        <p style="margin: 0 0 10px; color: #666666; font-size: 15px; line-height: 1.6;">
                                            • Keep using your email and password if you prefer
                                        </p>
                                        <p style="margin: 0; color: #666666; font-size: 15px; line-height: 1.6;">
                                            • Enhanced account security with multiple login options
                                        </p>
                                    </td>
                                </tr>
                            </table>
                            
                            <!-- Security notice -->
                            <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-top: 30px;">
                                <tr>
                                    <td style="background-color: #fff9f7; border-left: 4px solid #bc3a08; padding: 20px; border-radius: 8px;">
                                        <p style="margin: 0 0 8px; color: #bc3a08; font-size: 15px; font-weight: 700;">
                                            🔒 Security Alert
                                        </p>
                                        <p style="margin: 0; color: #666666; font-size: 14px; line-height: 1.6;">
                                            If you didn't authorize this account linking, please contact our support team immediately to secure your account.
                                        </p>
                                    </td>
                                </tr>
                            </table>
                            
                            <p style="margin: 30px 0 0; color: #999999; font-size: 14px; line-height: 1.6; text-align: center;">
                                Questions about this change? Our support team is always happy to help!
                            </p>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #fafafa; padding: 30px 40px; text-align: center; border-top: 1px solid #eeeeee;">
                            <p style="margin: 0 0 8px; color: #999999; font-size: 13px;">
                                © 2024 Schema Builder. All rights reserved.
                            </p>
                            <p style="margin: 0; color: #cccccc; font-size: 12px;">
                                Design better databases, faster.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
    `, firstName)

	return s.sendEmail(email, subject, body)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.User))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.User, s.config.Password)

	if err := d.DialAndSend(m); err != nil {
		s.log.Errorf("Failed to send email to %s: %v", to, err)
		s.logEmailToConsole(to, subject, body)
		return fmt.Errorf("failed to send email: %v", err)
	}

	s.log.Infof("Email sent successfully to %s", to)
	return nil
}

func (s *EmailService) logEmailToConsole(to, subject, body string) {
	fmt.Printf("\n===========================================\n")
	fmt.Printf("EMAIL DETAILS (Fallback - SMTP Failed)\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("===========================================\n\n")
}
