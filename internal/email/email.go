package email

import (
	"fmt"
	"log"
	"net/smtp"
	"sync"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	Username     string
	Password     string
	FromAddress  string
	FromName     string
}

type EmailService struct {
	config EmailConfig
	client *smtp.Client
	mu     sync.Mutex
}

var (
	globalEmailService *EmailService
	once              sync.Once
)

func NewEmailService(config EmailConfig) *EmailService {
	once.Do(func() {
		globalEmailService = &EmailService{
			config: config,
		}
	})
	return globalEmailService
}

func GetEmailService() *EmailService {
	return globalEmailService
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	if s.config.SMTPHost == "" {
		log.Printf("[EMAIL] SMTP not configured, skipping email to %s", to)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.config.FromName, s.config.FromAddress, to, subject, body))

	err := smtp.SendMail(addr, auth, s.config.FromAddress, []string{to}, msg)
	if err != nil {
		log.Printf("[EMAIL] Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("[EMAIL] Email sent to %s", to)
	return nil
}

// SendPurchaseConfirmation 发送购票确认邮件
func (s *EmailService) SendPurchaseConfirmation(to, username, eventTitle, orderNo string, amount float64) error {
	subject := "购票成功确认"
	body := fmt.Sprintf(`
<html>
<body>
<h2>购票成功！</h2>
<p>尊敬的 %s，您好！</p>
<p>您已成功购买 <strong>%s</strong> 的门票。</p>
<p>订单号：%s</p>
<p>支付金额：¥%.2f</p>
<p>请在活动开始前出示电子票二维码入场。</p>
<br>
<p>感谢您的购买！</p>
</body>
</html>
`, username, eventTitle, orderNo, amount)

	return s.SendEmail(to, subject, body)
}

// SendPaymentSuccess 发送支付成功邮件
func (s *EmailService) SendPaymentSuccess(to, username, eventTitle, orderNo string, amount float64) error {
	subject := "支付成功通知"
	body := fmt.Sprintf(`
<html>
<body>
<h2>支付成功</h2>
<p>尊敬的 %s，您好！</p>
<p>您已成功支付 <strong>%s</strong> 的门票费用。</p>
<p>订单号：%s</p>
<p>支付金额：¥%.2f</p>
<p>您的电子票已准备就绪，请在"我的票务"中查看。</p>
</body>
</html>
`, username, eventTitle, orderNo, amount)

	return s.SendEmail(to, subject, body)
}

// SendEventReminder 发送活动提醒邮件
func (s *EmailService) SendEventReminder(to, username, eventTitle, eventTime, location string) error {
	subject := "活动即将开始提醒"
	body := fmt.Sprintf(`
<html>
<body>
<h2>活动提醒</h2>
<p>尊敬的 %s，您好！</p>
<p>您购买的 <strong>%s</strong> 即将开始。</p>
<p>活动时间：%s</p>
<p>活动地点：%s</p>
<p>请提前到达现场，出示电子票二维码入场。</p>
</body>
</html>
`, username, eventTitle, eventTime, location)

	return s.SendEmail(to, subject, body)
}

// SendWaitlistNotification 发送等候名单通知邮件
func (s *EmailService) SendWaitlistNotification(to, username, eventTitle string) error {
	subject := "候补成功通知"
	body := fmt.Sprintf(`
<html>
<body>
<h2>候补成功！</h2>
<p>尊敬的 %s，您好！</p>
<p>您候补的 <strong>%s</strong> 已有票源。</p>
<p>请在 24 小时内完成购票，逾期将自动释放。</p>
<p><a href="http://localhost:3003/events">立即购票</a></p>
</body>
</html>
`, username, eventTitle)

	return s.SendEmail(to, subject, body)
}
