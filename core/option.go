package core

import (
	"sync"
)

// Option provides a global key-value store backed by the web_options database table.
// It caches values in memory to avoid repeated database queries.
type Option struct {
	mu     sync.RWMutex
	cache  map[string]string
	loaded bool
}

var optionInstance *Option
var optionOnce sync.Once

// GetOption returns the global Option singleton
func GetOption() *Option {
	optionOnce.Do(func() {
		optionInstance = &Option{
			cache: make(map[string]string),
		}
	})
	return optionInstance
}

// Get retrieves an option value by key. Returns the value or defaultVal if not found.
// On first call, loads all options from the database into memory.
func (o *Option) Get(key string, defaultVal ...string) string {
	o.mu.RLock()
	if o.loaded {
		val, ok := o.cache[key]
		o.mu.RUnlock()
		if ok {
			return val
		}
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	o.mu.RUnlock()

	// First call: load from database
	o.refresh()

	o.mu.RLock()
	val, ok := o.cache[key]
	o.mu.RUnlock()
	if ok {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// Set updates or inserts an option value in the database and refreshes the cache
func (o *Option) Set(key, value string) error {
	db := DB()
	// Use upsert: INSERT ... ON CONFLICT DO UPDATE
	result := db.Exec(
		"INSERT INTO web_options (name, value, created_at, updated_at) VALUES (?, ?, NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET value = ?, updated_at = NOW()",
		key, value, value,
	)
	if result.Error != nil {
		return result.Error
	}
	o.refresh()
	return nil
}

// SetMany batch-upserts multiple key-value pairs in a single transaction,
// then refreshes the cache only once — avoiding repeated queries on bulk saves.
func (o *Option) SetMany(options map[string]string) error {
	if len(options) == 0 {
		return nil
	}
	db := DB()
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for key, value := range options {
		result := tx.Exec(
			"INSERT INTO web_options (name, value, created_at, updated_at) VALUES (?, ?, NOW(), NOW()) ON CONFLICT (name) DO UPDATE SET value = ?, updated_at = NOW()",
			key, value, value,
		)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	o.refresh()
	return nil
}

// refresh reloads the cache from the database
func (o *Option) refresh() {
	db := DB()
	var options []struct {
		Name  string
		Value string
	}
	db.Table("web_options").Select("name", "value").Find(&options)

	o.mu.Lock()
	defer o.mu.Unlock()
	o.cache = make(map[string]string, len(options))
	for _, opt := range options {
		o.cache[opt.Name] = opt.Value
	}
	o.loaded = true
}

// AdminOption is a global shorthand for GetOption().Get()
func AdminOption(key string, defaultVal ...string) string {
	return GetOption().Get(key, defaultVal...)
}

// AdminOptionSet is a global shorthand for GetOption().Set()
func AdminOptionSet(key, value string) error {
	return GetOption().Set(key, value)
}

// AdminOptionSetMany is a global shorthand for GetOption().SetMany()
func AdminOptionSetMany(options map[string]string) error {
	return GetOption().SetMany(options)
}
