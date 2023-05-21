package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type Cache struct {
	CacheDir      string
	CacheFilePath string
	ConfigDir     string
	Users         Users
}

func NewCache() Cache {
	return Cache{}
}

func (c *Cache) GetCache() {
	// Check for cache data in $XDG_CACHE_HOME or $HOME/.cache.
	// If not found, download and cache
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("failed to get eventsub cache directory: %w", err)
	}

	cacheDir = filepath.Join(cacheDir, "subsleuth")
	if _, err := os.Stat(cacheDir); err != nil {
		log.Infof("cache directory not found, creating: %s", cacheDir)
		os.MkdirAll(cacheDir, 0755)
	}

	cacheFilePath := filepath.Join(cacheDir, "/subs.json")
	c.CacheDir, c.CacheFilePath = cacheDir, cacheFilePath
	if _, err := os.Stat(c.CacheFilePath); err == nil {
		// Cache file exists, no need to download
		// Parse json and return EventSubsLists
		cacheFile, err := os.Open(c.CacheFilePath)
		if err != nil {
			log.Warnf("failed to open cache file for reading: %s: %s", c.CacheFilePath, err)
		}
		defer cacheFile.Close()

		var eventSubsLists EventSubsLists
		json.NewDecoder(cacheFile).Decode(&eventSubsLists)
		if err != nil {
			log.Errorf("failed to decode repsonse: %s", err)
		}
	} else {
		// Cache file does not exist
		// download and cache by calling the Twitch CLI
		log.Info("downloading and caching from twitch")
		resp, err := RunTwitchCli([]string{"api", "get", "-P", "eventsub/subscriptions"})
		if err != nil {
			log.Fatalf("failed to download subscriptions: %s", resp)
		}
		eventSubs := EventSubsLists{}
		json.Unmarshal(resp, &eventSubs)
		log.Debug(resp)
		log.Debug(eventSubs)
		log.Debug(eventSubs.Subscription[0].Type)

		cacheFile, err := os.OpenFile(c.CacheFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Infof("failed to open cache file for writing: %s, %s", c.CacheFilePath, err)
		}
		defer cacheFile.Close()

		n, err := cacheFile.Write(resp)
		if err != nil {
			log.Fatalf("failed to write to cache file: %s, %s", c.CacheFilePath, err)
		}
		log.Debugf("wrote %d bytes to cache file", n)
	}
}

func (c *Cache) FlushUsers() {
	usersCacheFilePath := filepath.Join(c.CacheDir, "/users.json")
	userData, err := json.Marshal(c.Users)
	if err != nil {
		log.Fatalf("failed to marshal users data: %s", err)
	}
	os.WriteFile(usersCacheFilePath, userData, 0755)
}

func (c *Cache) LoadUsers() {
	usersCacheFilePath := filepath.Join(c.CacheDir, "/users.json")
	if _, err := os.Stat(usersCacheFilePath); err == nil {
		// Cache file exists, no need to download
		// Parse json and return EventSubsLists
		cacheFile, err := os.Open(usersCacheFilePath)
		if err != nil {
			log.Warnf("failed to open cache file for reading: %s: %s", usersCacheFilePath, err)
		}
		defer cacheFile.Close()

		var users Users
		json.NewDecoder(cacheFile).Decode(&users)
		if err != nil {
			log.Errorf("failed to decode repsonse: %s", err)
		}
		c.Users = users
	} else {
		c.Users = Users{}
	}
}

func (c Cache) GetUserById(id string) User {
	for _, user := range c.Users.Users {
		if user.ID == id {
			return user
		}
	}
	return User{}
}

func (c Cache) GetUserByName(name string) User {
	for _, user := range c.Users.Users {
		if user.Login == name {
			return user
		}
	}
	return User{}
}
