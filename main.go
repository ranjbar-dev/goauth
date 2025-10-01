package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pquerna/otp/totp"
	"gopkg.in/yaml.v3"
)

type Account struct {
	ID       int    `yaml:"id"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Site     string `yaml:"site"`
	Secret   string `yaml:"secret"`
}

type Config struct {
	Accounts []Account `yaml:"accounts"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func generateTOTP(secret string) (string, error) {
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", err
	}
	return code, nil
}

func getRemainingTime() int {
	now := time.Now().Unix()
	return 30 - int(now%30)
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func drawTable(accounts []Account) {
	clearScreen()

	// Colors
	borderColor := color.New(color.FgHiBlack)         // Gray borders
	headerColor := color.New(color.FgGreen)           // Green headers
	lightGrayColor := color.New(color.FgHiWhite)      // Light gray for id, name, site, username
	codeColor := color.New(color.FgWhite, color.Bold) // White for generated code
	timeColor := color.New(color.FgMagenta)           // Keep magenta for time

	remaining := getRemainingTime()

	// Calculate column widths
	idLen := 4       // "ID"
	maxNameLen := 6  // "Name"
	maxUserLen := 10 // "Username"
	maxSiteLen := 6  // "Site"
	for _, acc := range accounts {
		if len(fmt.Sprintf("%d", acc.ID)) > idLen-2 {
			idLen = len(fmt.Sprintf("%d", acc.ID)) + 2
		}
		if len(acc.Name) > maxNameLen {
			maxNameLen = len(acc.Name)
		}
		if len(acc.Username) > maxUserLen {
			maxUserLen = len(acc.Username)
		}
		if len(acc.Site) > maxSiteLen {
			maxSiteLen = len(acc.Site)
		}
	}

	// Add padding
	idLen += 2
	maxNameLen += 2
	maxUserLen += 2
	maxSiteLen += 2
	codeLen := 10 // TOTP code length + padding
	timeLen := 8  // "XXs" + padding

	totalWidth := idLen + maxNameLen + maxUserLen + maxSiteLen + codeLen + timeLen + 7 // 7 for separators

	// Draw top border
	borderColor.Println(strings.Repeat("═", totalWidth))

	// Draw header
	borderColor.Print("║ ")
	headerColor.Printf("%-*s", idLen, "ID")
	borderColor.Print("│ ")
	headerColor.Printf("%-*s", maxNameLen, "Name")
	borderColor.Print("│ ")
	headerColor.Printf("%-*s", maxUserLen, "Username")
	borderColor.Print("│ ")
	headerColor.Printf("%-*s", maxSiteLen, "Site")
	borderColor.Print("│ ")
	headerColor.Printf("%-*s", codeLen, "Code")
	borderColor.Print("│ ")
	headerColor.Printf("%-*s", timeLen, "Time")
	borderColor.Println(" ║")

	// Draw separator
	borderColor.Println(strings.Repeat("═", totalWidth))

	// Draw rows
	for _, acc := range accounts {
		code, err := generateTOTP(acc.Secret)
		if err != nil {
			code = "ERROR"
		}

		// Format code with space in middle (XXX XXX)
		if len(code) == 6 {
			code = code[:3] + " " + code[3:]
		}

		borderColor.Print("║ ")
		lightGrayColor.Printf("%-*d", idLen, acc.ID)
		borderColor.Print("│ ")
		lightGrayColor.Printf("%-*s", maxNameLen, acc.Name)
		borderColor.Print("│ ")
		lightGrayColor.Printf("%-*s", maxUserLen, acc.Username)
		borderColor.Print("│ ")
		lightGrayColor.Printf("%-*s", maxSiteLen, acc.Site)
		borderColor.Print("│ ")
		codeColor.Printf(" %-*s", codeLen-1, code)
		borderColor.Print("│ ")

		// Color time based on remaining seconds
		timeStr := fmt.Sprintf("%-ds", remaining)
		if remaining <= 5 {
			color.New(color.FgRed, color.Bold).Printf(" %-*s", timeLen-1, timeStr)
		} else if remaining <= 10 {
			color.New(color.FgYellow, color.Bold).Printf(" %-*s", timeLen-1, timeStr)
		} else {
			timeColor.Printf(" %-*s", timeLen-1, timeStr)
		}
		borderColor.Println(" ║")
	}

	// Draw bottom border
	borderColor.Println(strings.Repeat("═", totalWidth))

	// Draw footer with current time
	fmt.Print("\n")
	color.New(color.FgCyan).Printf("⟳ Auto-refresh | Current time: %s\n", time.Now().Format("15:04:05"))
	color.New(color.FgHiBlack).Println("Press Ctrl+C to exit")
}

func main() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if len(config.Accounts) == 0 {
		log.Fatal("No accounts found in config file")
	}

	// Initial draw
	drawTable(config.Accounts)

	// Refresh every second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		drawTable(config.Accounts)
	}
}
