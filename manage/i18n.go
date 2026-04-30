package manage

// ─── Translation system for PureCore Manager ───────────────

type Lang string

const (
	LangEN Lang = "en"
	LangZH Lang = "zh"
)

// Global language state
var CurrentLang Lang = LangEN

func T(key string) string {
	switch CurrentLang {
	case LangZH:
		if msg, ok := msgZH[key]; ok {
			return msg
		}
	default:
		if msg, ok := msgEN[key]; ok {
			return msg
		}
	}
	// Fallback to English if key not found
	if msg, ok := msgEN[key]; ok {
		return msg
	}
	return key
}

// ToggleLang switches between EN and ZH
func ToggleLang() Lang {
	if CurrentLang == LangEN {
		CurrentLang = LangZH
	} else {
		CurrentLang = LangEN
	}
	return CurrentLang
}

// ─── English messages ─────────────────────────────────────
var msgEN = map[string]string{
	"title":                "PureCore Management",
	"title_sub":            "Manage your PureCore instance",
	"loading":              "Loading...",
	"checking":             "Checking PureCore installation...",
	"compose_found":        "docker-compose.yml found",
	"env_found":            ".env file found",
	"containers_found":     "Running containers: %d",
	"installed_ok":         "PureCore installation detected ✓",
	"installed_no":         "PureCore Not Detected",
	"no_install_desc":      "This directory does not appear to contain a valid PureCore installation.",
	"no_install_hint":      "Run the deploy script first: ./scripts/deploy.sh",
	"not_installed_exit":   "Not a PureCore project directory. Exiting.",
	"menu_title":           "Management Menu",
	"menu_switch_prefix":   "Switch Admin Route Prefix",
	"menu_switch_theme":    "Switch Frontend Theme",
	"menu_update":          "Update PureCore Version",
	"menu_status":          "View Service Status",
	"menu_restart":         "Restart Services",
	"menu_logs":            "View Service Logs",
	"menu_shell":           "Open Shell in Container",
	"menu_lang":            "Switch Language (EN/ZH)",
	"menu_exit":            "Exit",
	"current_prefix":       "Current prefix: %s",
	"enter_new_prefix":     "Enter new admin route prefix:",
	"prefix_updated":       "Admin route prefix updated to: %s",
	"restart_question":     "Restart services to apply the change? (y/N): ",
	"restarting":           "Restarting services...",
	"restart_done":         "Services restarted successfully.",
	"restart_skipped":      "Restart skipped. Change will take effect after manual restart.",
	"current_theme":        "Current theme: %s",
	"select_theme":         "Select DaisyUI Theme",
	"theme_updated":        "Theme updated to: %s",
	"update_title":         "Update PureCore",
	"current_version":      "Current version: %s",
	"fetching_versions":    "Fetching available versions from GitHub...",
	"select_version":       "Select Version",
	"manual_input_option":  "[Manual input]",
	"enter_version_manual": "Enter version (e.g., 1.0.7, latest): ",
	"version_selected":     "Selected: %s",
	"updating":             "Pulling images and restarting (version: %s)...",
	"update_done":          "Update complete! Now running version %s.",
	"update_failed":        "Update failed. Check docker compose logs.",
	"status_title":         "Service Status",
	"status_version":       "Version",
	"status_prefix":        "Admin Prefix",
	"status_theme":         "Theme",
	"status_frontend":      "Frontend URL",
	"status_admin":         "Admin URL",
	"no_containers":        "No PureCore containers are running.",
	"restarting_title":     "Restarting Services",
	"restart_ok":           "Services restarted successfully ✓",
	"restart_fail":         "Failed to restart services.",
	"logs_title":           "View Logs",
	"select_service":       "Select service:",
	"logs_all":             "All Services",
	"logs_backend":         "Backend",
	"logs_frontend":        "Frontend",
	"logs_database":        "Database",
	"shell_title":          "Open Shell",
	"select_container":     "Select container:",
	"shell_backend":        "Backend (purecore-backend)",
	"shell_frontend":       "Frontend (purecore-frontend)",
	"shell_database":       "Database (purecore-db)",
	"shell_entering":       "Opening shell in %s...",
	"shell_note":           "Run this command in your terminal:",
	"goodbye":              "Goodbye!",
	"press_enter":          "Press Enter to continue...",
	"confirm_yes":          "y",
	"confirm_no":           "N",
	"daisyui_themes":       "Available DaisyUI themes",
	"navigate_hint":        "Use ↑/↓ arrows to navigate, Enter to select, Esc to cancel",
	"back":                 "← Back",
	"loading_failed":       "Failed to load data.",
	"docker_not_found":     "Docker is not available. Please install Docker first.",
	"compose_not_found":    "docker compose command not found.",
}

// ─── Chinese messages ─────────────────────────────────────
var msgZH = map[string]string{
	"title":                "PureCore 管理工具",
	"title_sub":            "管理你的 PureCore 实例",
	"loading":              "加载中...",
	"checking":             "正在检查 PureCore 安装状态...",
	"compose_found":        "找到 docker-compose.yml",
	"env_found":            "找到 .env 文件",
	"containers_found":     "运行中的容器: %d",
	"installed_ok":         "已检测到 PureCore 安装 ✓",
	"installed_no":         "未检测到 PureCore",
	"no_install_desc":      "当前目录不包含有效的 PureCore 安装。",
	"no_install_hint":      "请先运行部署脚本: ./scripts/deploy.sh",
	"not_installed_exit":   "不是 PureCore 项目目录，退出。",
	"menu_title":           "管理菜单",
	"menu_switch_prefix":   "切换后台路由前缀",
	"menu_switch_theme":    "切换前端默认主题",
	"menu_update":          "更新 PureCore 版本",
	"menu_status":          "查看服务状态",
	"menu_restart":         "重启服务",
	"menu_logs":            "查看服务日志",
	"menu_shell":           "进入容器 Shell",
	"menu_lang":            "切换语言 (EN/ZH)",
	"menu_exit":            "退出",
	"current_prefix":       "当前前缀: %s",
	"enter_new_prefix":     "输入新的后台路由前缀:",
	"prefix_updated":       "后台路由前缀已更新为: %s",
	"restart_question":     "是否重启服务以应用更改？(y/N): ",
	"restarting":           "正在重启服务...",
	"restart_done":         "服务已成功重启。",
	"restart_skipped":      "已跳过重启。更改将在手动重启后生效。",
	"current_theme":        "当前主题: %s",
	"select_theme":         "选择 DaisyUI 主题",
	"theme_updated":        "主题已更新为: %s",
	"update_title":         "更新 PureCore",
	"current_version":      "当前版本: %s",
	"fetching_versions":    "正在从 GitHub 获取可用版本...",
	"select_version":       "选择版本",
	"manual_input_option":  "[手动输入]",
	"enter_version_manual": "输入版本号（例如: 1.0.7, latest）: ",
	"version_selected":     "已选择: %s",
	"updating":             "正在拉取镜像并重启（版本: %s）...",
	"update_done":          "更新完成！现在运行的是 %s 版本。",
	"update_failed":        "更新失败。请检查 docker compose logs。",
	"status_title":         "服务状态",
	"status_version":       "版本",
	"status_prefix":        "后台前缀",
	"status_theme":         "主题",
	"status_frontend":      "前端地址",
	"status_admin":         "后台地址",
	"no_containers":        "没有运行中的 PureCore 容器。",
	"restarting_title":     "正在重启服务",
	"restart_ok":           "服务已成功重启 ✓",
	"restart_fail":         "服务重启失败。",
	"logs_title":           "查看日志",
	"select_service":       "选择服务:",
	"logs_all":             "全部服务",
	"logs_backend":         "后端",
	"logs_frontend":        "前端",
	"logs_database":        "数据库",
	"shell_title":          "打开 Shell",
	"select_container":     "选择容器:",
	"shell_backend":        "后端 (purecore-backend)",
	"shell_frontend":       "前端 (purecore-frontend)",
	"shell_database":       "数据库 (purecore-db)",
	"shell_entering":       "正在进入 %s...",
	"shell_note":           "请在终端中运行以下命令:",
	"goodbye":              "再见！",
	"press_enter":          "按 Enter 继续...",
	"confirm_yes":          "y",
	"confirm_no":           "N",
	"daisyui_themes":       "可用的 DaisyUI 主题",
	"navigate_hint":        "使用 ↑/↓ 方向键导航，Enter 确认，Esc 取消",
	"back":                 "← 返回",
	"loading_failed":       "加载数据失败。",
	"docker_not_found":     "Docker 不可用。请先安装 Docker。",
	"compose_not_found":    "未找到 docker compose 命令。",
}
