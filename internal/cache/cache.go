package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultTTL = 24 * time.Hour
)

var CacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "pulp")

type CacheEntry struct {
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	TTLDuration int64   `json:"ttl_duration"`
}

func urlHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:16]
}

func CachePath() string {
	return CacheDir
}

func ensureCacheDir() error {
	return os.MkdirAll(CacheDir, 0755)
}

func Get(url string) (string, error) {
	if err := ensureCacheDir(); err != nil {
		return "", err
	}

	hash := urlHash(url)
	path := filepath.Join(CacheDir, hash+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("cache miss")
		}
		return "", err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", err
	}

	ttl := time.Duration(entry.TTLDuration) * time.Nanosecond
	if ttl == 0 {
		ttl = DefaultTTL
	}

	if time.Since(entry.Timestamp) > ttl {
		return "", fmt.Errorf("cache expired")
	}

	return entry.Content, nil
}

func Set(url, content string, ttl time.Duration) error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	hash := urlHash(url)
	path := filepath.Join(CacheDir, hash+".json")

	if ttl == 0 {
		ttl = DefaultTTL
	}

	entry := CacheEntry{
		Content:    content,
		Timestamp:  time.Now(),
		TTLDuration: int64(ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Invalidate(url string) error {
	hash := urlHash(url)
	path := filepath.Join(CacheDir, hash+".json")
	return os.Remove(path)
}

func ClearAll() error {
	if err := ensureCacheDir(); err != nil {
		return err
	}

	entries, err := os.ReadDir(CacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := os.Remove(filepath.Join(CacheDir, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}