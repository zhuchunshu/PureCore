package controllers

import (
	models "purecore/app/Models"
	"purecore/core"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// parseUserAgent extracts device and browser info from a User-Agent string
func parseUserAgent(ua string) (deviceType, deviceBrand, deviceModel, browser, os string) {
	uaLower := strings.ToLower(ua)

	// Detect OS
	switch {
	case strings.Contains(uaLower, "windows nt 10"):
		os = "Windows 10"
	case strings.Contains(uaLower, "windows nt 6.3"):
		os = "Windows 8.1"
	case strings.Contains(uaLower, "windows nt 6.2"):
		os = "Windows 8"
	case strings.Contains(uaLower, "windows nt 6.1"):
		os = "Windows 7"
	case strings.Contains(uaLower, "windows nt"):
		os = "Windows"
	case strings.Contains(uaLower, "mac os x") || strings.Contains(uaLower, "macintosh"):
		os = "macOS"
	case strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipad"):
		os = "iOS"
	case strings.Contains(uaLower, "android"):
		os = "Android"
	case strings.Contains(uaLower, "linux"):
		os = "Linux"
	case strings.Contains(uaLower, "cros"):
		os = "ChromeOS"
	default:
		os = ""
	}

	// Detect Browser
	switch {
	case strings.Contains(uaLower, "edg"):
		browser = "Edge"
	case strings.Contains(uaLower, "chrome"):
		browser = "Chrome"
	case strings.Contains(uaLower, "firefox"):
		browser = "Firefox"
	case strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome"):
		browser = "Safari"
	case strings.Contains(uaLower, "opera") || strings.Contains(uaLower, "opr"):
		browser = "Opera"
	case strings.Contains(ua, "MSIE") || strings.Contains(ua, "Trident"):
		browser = "Internet Explorer"
	default:
		browser = ""
	}

	// Detect device type and model
	mobile := strings.Contains(uaLower, "mobi") ||
		strings.Contains(uaLower, "android") && strings.Contains(uaLower, "mobile") ||
		strings.Contains(uaLower, "iphone")

	tablet := strings.Contains(uaLower, "ipad") ||
		(strings.Contains(uaLower, "android") && !strings.Contains(uaLower, "mobile"))

	switch {
	case tablet:
		deviceType = "tablet"
	case mobile:
		deviceType = "mobile"
	default:
		deviceType = "desktop"
	}

	// Extract brand and model for mobile devices
	if deviceType == "mobile" || deviceType == "tablet" {
		if strings.Contains(uaLower, "iphone") {
			deviceBrand = "Apple"
			deviceModel = extractiPhoneModel(uaLower)
		} else if strings.Contains(uaLower, "ipad") {
			deviceBrand = "Apple"
			deviceModel = "iPad"
		} else if strings.Contains(uaLower, "samsung") || strings.Contains(uaLower, "sm-") {
			deviceBrand = "Samsung"
			deviceModel = extractAndroidModel(uaLower)
		} else if strings.Contains(uaLower, "pixel") {
			deviceBrand = "Google"
			deviceModel = "Pixel"
		} else if strings.Contains(uaLower, "huawei") || strings.Contains(uaLower, "honor") {
			deviceBrand = "Huawei"
			deviceModel = extractAndroidModel(uaLower)
		} else if strings.Contains(uaLower, "xiaomi") || strings.Contains(uaLower, "redmi") || strings.Contains(uaLower, "poco") {
			deviceBrand = "Xiaomi"
			deviceModel = extractAndroidModel(uaLower)
		} else if strings.Contains(uaLower, "oppo") {
			deviceBrand = "OPPO"
			deviceModel = extractAndroidModel(uaLower)
		} else if strings.Contains(uaLower, "vivo") {
			deviceBrand = "vivo"
			deviceModel = extractAndroidModel(uaLower)
		} else if strings.Contains(uaLower, "oneplus") {
			deviceBrand = "OnePlus"
			deviceModel = extractAndroidModel(uaLower)
		} else {
			deviceModel = extractAndroidModel(uaLower)
		}
	}

	return
}

// extractiPhoneModel extracts iPhone model from user agent
func extractiPhoneModel(ua string) string {
	// iPhone models like "iPhone15,2" → "iPhone 14 Pro" mappings
	re := regexp.MustCompile(`iphone(\d+),(\d+)`)
	matches := re.FindStringSubmatch(ua)
	if len(matches) > 0 {
		// Simplified mapping — return generic "iPhone" with chip generation
		return "iPhone"
	}
	// Try "iPhone X" style
	re2 := regexp.MustCompile(`(?i)iphone\s+(\d+[a-z]*)`)
	matches2 := re2.FindStringSubmatch(ua)
	if len(matches2) > 1 {
		return "iPhone " + strings.ToUpper(matches2[1])
	}
	return "iPhone"
}

// extractAndroidModel extracts Android device model from user agent
func extractAndroidModel(ua string) string {
	// Pattern: "; SM-G991B Build/" or "; Pixel 7 Build/" etc.
	re := regexp.MustCompile(`;\s*([\w\s-]+?)\s+Build/`)
	matches := re.FindStringSubmatch(ua)
	if len(matches) > 1 {
		model := strings.TrimSpace(matches[1])
		// Filter out generic tokens
		if !contains([]string{"mobile", "linux", "android", "applewebkit", "khtml", "gecko", "safari", "chrome", "version"}, strings.ToLower(model)) {
			return model
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// UserSessionController handles user session management
type UserSessionController struct{}

// Index lists all active sessions for the current user
func (usc *UserSessionController) Index(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	var sessions []models.UserSession
	if err := core.DB().Where("user_id = ?", userID).
		Order("is_current DESC, last_activity DESC").
		Find(&sessions).Error; err != nil {
		return res.Error("Failed to fetch sessions", 500)
	}

	return res.Success(sessions)
}

// Revoke revokes a specific session
func (usc *UserSessionController) Revoke(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	sessionID := req.Input("id")
	if sessionID == "" {
		return res.Error("Session ID is required", 422)
	}

	var session models.UserSession
	if err := core.DB().Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return res.NotFound("Session not found")
	}

	if err := core.DB().Delete(&session).Error; err != nil {
		return res.Error("Failed to revoke session", 500)
	}

	return res.Success(nil)
}

// RevokeAll revokes all sessions except the current one
func (usc *UserSessionController) RevokeAll(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	// Delete all sessions for this user that are not marked as current
	if err := core.DB().Where("user_id = ? AND is_current = ?", userID, false).Delete(&models.UserSession{}).Error; err != nil {
		return res.Error("Failed to revoke sessions", 500)
	}

	return res.Success(nil)
}

// CreateSession creates a new session record (called during login/register)
func CreateSession(c fiber.Ctx, userID uint) (*models.UserSession, error) {
	ua := c.Get("User-Agent")
	ip := c.IP()

	deviceType, deviceBrand, deviceModel, browser, os := parseUserAgent(ua)

	sessionToken := uuid.New().String()

	session := models.UserSession{
		UserID:       userID,
		IPAddress:    ip,
		UserAgent:    ua,
		DeviceType:   deviceType,
		DeviceBrand:  deviceBrand,
		DeviceModel:  deviceModel,
		Browser:      browser,
		OS:           os,
		SessionToken: sessionToken,
		IsCurrent:    true,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := core.DB().Create(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}
