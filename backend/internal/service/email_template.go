package service

import (
	"fmt"
	"time"

	"github.com/matcornic/hermes/v2"
)

// getEmailHermes 获取配置好的 hermes 实例
func getEmailHermes() hermes.Hermes {
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	return hermes.Hermes{
		Theme: new(hermes.Default),
		Product: hermes.Product{
			Name:        "FRP Panel",
			Link:        panelURL,
			Copyright:   "© FRP Panel - 内网穿透管理平台",
			TroubleText: "如果按钮无法点击，请复制以下链接到浏览器打开：",
		},
	}
}

// TrafficAlertData 流量告警数据
type TrafficAlertData struct {
	ProxyName    string
	AlertType    string
	CurrentValue string
	Threshold    string
	Time         time.Time
}

// OfflineAlertData 离线告警数据
type OfflineAlertData struct {
	TargetType string
	TargetName string
	Message    string
	Time       time.Time
}

// RecoveryAlertData 恢复通知数据
type RecoveryAlertData struct {
	TargetType string
	TargetName string
	Time       time.Time
}

// TestEmailData 测试邮件数据
type TestEmailData struct {
	Host string
	Port string
	SSL  bool
}

// SystemAlertData 系统告警数据
type SystemAlertData struct {
	AlertType string
	Message   string
	EventData string
	Time      time.Time
}

// GenerateTrafficAlertEmail 生成流量告警邮件
func GenerateTrafficAlertEmail(data TrafficAlertData) (html, text string, err error) {
	h := getEmailHermes()
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	email := hermes.Email{
		Body: hermes.Body{
			Title: "流量告警通知",
			Intros: []string{
				"您的代理流量已超过设定阈值，请及时处理。",
			},
			Dictionary: []hermes.Entry{
				{Key: "代理名称", Value: data.ProxyName},
				{Key: "告警类型", Value: data.AlertType},
				{Key: "当前值", Value: data.CurrentValue},
				{Key: "阈值", Value: data.Threshold},
				{Key: "时间", Value: data.Time.Format("2006-01-02 15:04:05")},
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击下方按钮查看详情：",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "查看详情",
						Link:  panelURL,
					},
				},
			},
			Outros: []string{
				"请及时处理以避免服务中断。",
			},
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("生成HTML邮件失败: %w", err)
	}
	text, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("生成纯文本邮件失败: %w", err)
	}
	return html, text, nil
}

// GenerateOfflineAlertEmail 生成离线告警邮件
func GenerateOfflineAlertEmail(data OfflineAlertData) (html, text string, err error) {
	h := getEmailHermes()
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	email := hermes.Email{
		Body: hermes.Body{
			Title: "离线告警通知",
			Intros: []string{
				fmt.Sprintf("您的 %s 已离线，请及时检查。", data.TargetType),
			},
			Dictionary: []hermes.Entry{
				{Key: "目标类型", Value: data.TargetType},
				{Key: "目标名称", Value: data.TargetName},
				{Key: "告警消息", Value: data.Message},
				{Key: "时间", Value: data.Time.Format("2006-01-02 15:04:05")},
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击下方按钮立即检查：",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "立即检查",
						Link:  panelURL,
					},
				},
			},
			Outros: []string{
				"请及时处理以恢复服务。",
			},
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("生成HTML邮件失败: %w", err)
	}
	text, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("生成纯文本邮件失败: %w", err)
	}
	return html, text, nil
}

// GenerateSystemAlertEmail 生成系统告警邮件
func GenerateSystemAlertEmail(data SystemAlertData) (html, text string, err error) {
	h := getEmailHermes()
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	email := hermes.Email{
		Body: hermes.Body{
			Title: "系统告警通知",
			Intros: []string{
				fmt.Sprintf("系统事件: %s", data.AlertType),
			},
			Dictionary: []hermes.Entry{
				{Key: "告警类型", Value: data.AlertType},
				{Key: "告警消息", Value: data.Message},
				{Key: "时间", Value: data.Time.Format("2006-01-02 15:04:05")},
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击下方按钮查看详情：",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "查看详情",
						Link:  panelURL,
					},
				},
			},
			Outros: []string{
				"请及时处理相关事项。",
			},
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("生成HTML邮件失败: %w", err)
	}
	text, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("生成纯文本邮件失败: %w", err)
	}
	return html, text, nil
}

// GenerateRecoveryAlertEmail 生成恢复通知邮件
func GenerateRecoveryAlertEmail(data RecoveryAlertData) (html, text string, err error) {
	h := getEmailHermes()
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	email := hermes.Email{
		Body: hermes.Body{
			Title: "恢复通知",
			Intros: []string{
				fmt.Sprintf("您的 %s 已恢复在线。", data.TargetType),
			},
			Dictionary: []hermes.Entry{
				{Key: "目标类型", Value: data.TargetType},
				{Key: "目标名称", Value: data.TargetName},
				{Key: "恢复时间", Value: data.Time.Format("2006-01-02 15:04:05")},
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击下方按钮查看状态：",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "查看状态",
						Link:  panelURL,
					},
				},
			},
			Outros: []string{
				"服务已恢复正常运行。",
			},
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("生成HTML邮件失败: %w", err)
	}
	text, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("生成纯文本邮件失败: %w", err)
	}
	return html, text, nil
}

// GenerateTestEmail 生成测试邮件
func GenerateTestEmail(data TestEmailData) (html, text string, err error) {
	h := getEmailHermes()
	settingService := NewSettingService()
	panelURL := settingService.GetPanelURL()

	sslText := "否"
	if data.SSL {
		sslText = "是"
	}
	email := hermes.Email{
		Body: hermes.Body{
			Title: "邮件配置测试",
			Intros: []string{
				"这是一封测试邮件，用于验证 FRP Panel 的邮件配置是否正确。",
			},
			Dictionary: []hermes.Entry{
				{Key: "SMTP 服务器", Value: data.Host},
				{Key: "端口", Value: data.Port},
				{Key: "SSL/TLS", Value: sslText},
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击下方按钮打开面板：",
					Button: hermes.Button{
						Color: "#3869D4",
						Text:  "打开 FRP Panel",
						Link:  panelURL,
					},
				},
			},
			Outros: []string{
				"如果您收到此邮件，说明邮件配置正确。",
			},
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("生成HTML邮件失败: %w", err)
	}
	text, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("生成纯文本邮件失败: %w", err)
	}
	return html, text, nil
}
