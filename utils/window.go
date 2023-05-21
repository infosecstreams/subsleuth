package utils

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GetWindowSize returns the width and height of the terminal window
func GetWindowSize() (width int, height int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, _ := cmd.Output()

	size := strings.Split(string(out), " ")
	width, _ = strconv.Atoi(strings.TrimSpace(size[1]))
	height, _ = strconv.Atoi(strings.TrimSpace(size[0]))
	return
}
