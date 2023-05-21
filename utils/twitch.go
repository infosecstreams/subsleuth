package utils

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"

	log "github.com/charmbracelet/log"
)

// Check for the binary (/usr/local/bin/twitch)
func GetTwitchCliPath() (string, bool) {
	var twitchCli string
	var err error
	if os.Getenv("EVENTSUB_TWITCH_CLI_PATH") != "" {
		twitchCli = os.Getenv("EVENTSUB_TWITCH_CLI_PATH")
	} else {
		twitchCli, err = exec.LookPath("/usr/local/bin/twitch")
	}

	// Display bubbletea error modal
	if err != nil {
		log.Fatal("twitch cli not found")
		return "", false
	}
	// Check for ~/.config/twitch-cli/.twitch-cli.env and return warning if not found
	// and hint to run `twitch login`
	configDir, _ := os.UserConfigDir()
	if _, err := os.Stat(configDir + "/twitch-cli/.twitch-cli.env"); err != nil {
		log.Warn(err)
		log.Warn("twitch cli config not found")
		log.Warn("run `twitch login` to setup to twitch cli")
	}

	log.Debugf("twitch cli '%s' found in $PATH'", twitchCli)
	log.Debug("twitch cli seems configured")
	return twitchCli, true
}

func RunTwitchCli(args []string) ([]byte, error) {
	if twitchCli, installed := GetTwitchCliPath(); installed {
		cmd := exec.Command(twitchCli, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("failed to run twitch cli: %s", err)
			return nil, err
		}
		return out, nil
	}
	return nil, errors.New("twitch cli is not installed")
}

type User struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount       int    `json:"view_count"`
	Email           string `json:"email"`
}

type UserResponse struct {
	User []User `json:"data"`
}

type Users struct {
	Users []User
}

func (c *Cache) GetUsernames(broadcasterID string, ids ...string) string {
	// Check for the user in the cache first
	for _, user := range c.Users.Users {
		if user.ID == broadcasterID {
			return user.DisplayName
		}
	}
	// twitch api get users -q login="${1}"
	command := []string{"api", "get", "users", "-q", "id=" + broadcasterID}
	if len(ids) > 0 {
		command = append(command, "-q")
		for _, id := range ids {
			command = append(command, "id="+id)
		}
	}
	resp, err := RunTwitchCli(command)
	if err != nil {
		log.Errorf("failed to get broadcaster username: %s", err)
		return "👾"
	}
	var newUser UserResponse
	json.Unmarshal(resp, &newUser)
	c.Users.Users = append(c.Users.Users, newUser.User...)
	c.FlushUsers()
	return newUser.User[0].DisplayName
}
