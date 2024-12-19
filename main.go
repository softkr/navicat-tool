package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:noinline
func main() {
	if version, err := getNavicatVersion(); err != nil {
		fmt.Printf("Error getting version: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("Detected Navicat Premium version %s\n", version)
		if prefsFile := getPreferencesPath(version); prefsFile == "" {
			fmt.Printf("Version '%s' not handled\n", version)
			os.Exit(1)
		} else {
			fmt.Println("Resetting trial time...")
			if err := processPreferences(prefsFile); err != nil {
				fmt.Printf("Error processing preferences: %v\n", err)
				os.Exit(1)
			}
			if err := processAppSupport(); err != nil {
				fmt.Printf("Error processing app support: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Done")
		}
	}
}

func getNavicatVersion() (string, error) {
	if output, err := exec.Command("defaults", "read", "/Applications/Navicat Premium.app/Contents/Info.plist", "CFBundleShortVersionString").Output(); err != nil {
		return "", err
	} else {
		version := strings.TrimSpace(string(output))
		if matches := regexp.MustCompile(`^([^\.]+)`).FindStringSubmatch(version); len(matches) < 2 {
			return "", fmt.Errorf("unable to parse version: %s", version)
		} else {
			return matches[1], nil
		}
	}
}

func getPreferencesPath(version string) string {
	home := os.Getenv("HOME")
	switch version {
	case "17", "16":
		return filepath.Join(home, "Library/Preferences/com.navicat.NavicatPremium.plist")
	case "15":
		return filepath.Join(home, "Library/Preferences/com.prect.NavicatPremium15.plist")
	default:
		return ""
	}
}

func processPreferences(prefsFile string) error {
	if output, err := exec.Command("defaults", "read", prefsFile).Output(); err != nil {
		return err
	} else {
		if matches := regexp.MustCompile(`([0-9A-Z]{32}) =`).FindStringSubmatch(string(output)); len(matches) >= 2 {
			hash := matches[1]
			fmt.Printf("deleting %s array...\n", hash)
			return exec.Command("defaults", "delete", prefsFile, hash).Run()
		}
		return nil
	}
}

func processAppSupport() error {
	appSupportPath := filepath.Join(os.Getenv("HOME"), "Library/Application Support/PremiumSoft CyberTech/Navicat CC/Navicat Premium")
	
	if entries, err := os.ReadDir(appSupportPath); err != nil {
		return err
	} else {
		re := regexp.MustCompile(`\.([0-9A-Z]{32})`)
		for _, entry := range entries {
			if name := entry.Name(); strings.HasPrefix(name, ".") {
				if matches := re.FindStringSubmatch(name); len(matches) >= 2 {
					hash := matches[1]
					fmt.Printf("deleting %s folder...\n", hash)
					if err := os.RemoveAll(filepath.Join(appSupportPath, name)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}