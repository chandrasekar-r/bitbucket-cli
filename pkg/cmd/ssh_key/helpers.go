package sshkeycmd

import (
	"fmt"
	"os"
	"strings"
)

func readKeyMaterial(key, keyFile string) (string, error) {
	switch {
	case key != "" && keyFile != "":
		return "", fmt.Errorf("only one of --key or --key-file may be set")
	case key != "":
		return strings.TrimSpace(key), nil
	case keyFile != "":
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return "", fmt.Errorf("reading key file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	default:
		return "", fmt.Errorf("one of --key or --key-file is required")
	}
}