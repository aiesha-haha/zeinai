package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store holds the loaded translations: map[lang]map[key]value
type Store struct {
	translations map[string]map[string]string
	mu           sync.RWMutex
	defaultLang  string
}

var GlobalStore *Store

// Init initializes the i18n store by loading JSON files from the specified directory
func Init(localesPath string, defaultLang string) error {
	store := &Store{
		translations: make(map[string]map[string]string),
		defaultLang:  defaultLang,
	}

	files, err := os.ReadDir(localesPath)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		lang := file.Name()[0 : len(file.Name())-5] // remove .json
		content, err := os.ReadFile(filepath.Join(localesPath, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file.Name(), err)
		}

		var rawData map[string]interface{}
		if err := json.Unmarshal(content, &rawData); err != nil {
			return fmt.Errorf("failed to parse json %s: %w", file.Name(), err)
		}

		flatData := make(map[string]string)
		flatten("", rawData, flatData)

		store.translations[lang] = flatData
	}

	GlobalStore = store
	return nil
}

func flatten(prefix string, src map[string]interface{}, dest map[string]string) {
	for k, v := range src {
		newKey := k
		if prefix != "" {
			newKey = prefix + "." + k
		}
		switch child := v.(type) {
		case map[string]interface{}:
			flatten(newKey, child, dest)
		case string:
			dest[newKey] = child
		default:
			dest[newKey] = fmt.Sprintf("%v", child)
		}
	}
}

// Get returns the translation for a given key and language.
// Falls back to default language if not found.
func Get(lang, key string) string {
	if GlobalStore == nil {
		return key
	}
	GlobalStore.mu.RLock() // Use RLock for reading
	defer GlobalStore.mu.RUnlock()

	// 1. Try requested language
	if trans, ok := GlobalStore.translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}

	// 2. Try default language
	if lang != GlobalStore.defaultLang {
		if trans, ok := GlobalStore.translations[GlobalStore.defaultLang]; ok {
			if val, ok := trans[key]; ok {
				return val
			}
		}
	}

	// 3. Return key as fallback
	return key
}
