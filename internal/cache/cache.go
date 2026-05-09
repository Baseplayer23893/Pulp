package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	DefaultTTL = 24 * time.Hour
	MaxCacheItems = 100
	MaxCacheSizeMB = 50
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

// cleanupCache removes oldest entries if cache exceeds limits
func cleanupCache() error {
	entries, err := os.ReadDir(CacheDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	type fileInfo struct {
		name    string
		size    int64
		modTime time.Time
	}
	var files []fileInfo
	var totalSize int64
	for _, e := range entries {
		if info, err := os.Stat(filepath.Join(CacheDir, e.Name())); err == nil {
			files = append(files, fileInfo{e.Name(), info.Size(), info.ModTime()})
			totalSize += info.Size()
		}
	}

	maxSizeBytes := int64(MaxCacheSizeMB) * 1024 * 1024
	needsCleanup := len(files) > MaxCacheItems || totalSize > maxSizeBytes

	if !needsCleanup {
		return nil
	}

	// Sort by oldest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove oldest entries until under both limits
	for i := 0; i < len(files); i++ {
		removed := files[i].size
		if err := os.Remove(filepath.Join(CacheDir, files[i].name)); err != nil {
			continue
		}
		totalSize -= removed
		remaining := len(files) - i - 1
		if remaining <= MaxCacheItems && totalSize <= maxSizeBytes {
			break
		}
	}

	return nil
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

	// Cleanup before adding new entry
	cleanupCache()

	hash := urlHash(url)
	path := filepath.Join(CacheDir, hash+".json")

	if ttl == 0 {
		ttl = DefaultTTL
	}

	entry := CacheEntry{
		Content:     content,
		Timestamp:   time.Now(),
		TTLDuration: int64(ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
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