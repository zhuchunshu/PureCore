package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Lang 语言管理器，加载 lang/ 目录下的 JSON 翻译文件
type Lang struct {
	mu           sync.RWMutex
	locale       string
	translations map[string]map[string]string
	fallback     string
}

var langInstance *Lang
var langOnce sync.Once

// GetLang 获取全局语言实例（单例）
func GetLang() *Lang {
	langOnce.Do(func() {
		langInstance = &Lang{
			locale:       "zh",
			fallback:     "zh",
			translations: make(map[string]map[string]string),
		}
	})
	return langInstance
}

// InitLang 初始化语言，加载指定目录下的所有 JSON 文件
func InitLang(langDir string) error {
	l := GetLang()
	entries, err := os.ReadDir(langDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		locale := strings.TrimSuffix(entry.Name(), ".json")
		filePath := filepath.Join(langDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var flat map[string]string
		if err := json.Unmarshal(data, &flat); err != nil {
			// 尝试解析嵌套结构
			var nested map[string]map[string]string
			if err2 := json.Unmarshal(data, &nested); err2 != nil {
				continue
			}
			flat = make(map[string]string)
			for group, msgs := range nested {
				for key, val := range msgs {
					flat[group+"."+key] = val
				}
			}
		}
		l.mu.Lock()
		l.translations[locale] = flat
		l.mu.Unlock()
	}
	return nil
}

// SetLocale 设置当前语言
func (l *Lang) SetLocale(locale string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.locale = locale
}

// GetLocale 获取当前语言
func (l *Lang) GetLocale() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.locale
}

// Trans 翻译指定 key，支持点号分隔的嵌套 key（如 "common.success"）
// 如果找不到翻译则返回 fallback 语言的翻译，都找不到则返回 key 本身
func (l *Lang) Trans(key string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 尝试当前语言
	if msg, ok := l.getMsg(l.locale, key); ok {
		return msg
	}
	// 尝试回退语言
	if l.fallback != l.locale {
		if msg, ok := l.getMsg(l.fallback, key); ok {
			return msg
		}
	}
	return key
}

func (l *Lang) getMsg(locale, key string) (string, bool) {
	if msgs, ok := l.translations[locale]; ok {
		if msg, ok := msgs[key]; ok {
			return msg, true
		}
	}
	return "", false
}
