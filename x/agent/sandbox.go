package agent

import (
	"os"
	"path/filepath"
	"strings"
)

// Dangerous commands and path patterns to block.
var denyCommands = []string{
	"rm -rf", "rm -fr", "mkfs", "dd if=", "dd of=", "shred",
	"> /dev/", ">/dev/", "sudo ", "su ", "doas ", "chmod 777",
	"chown ", "chgrp ", "curl -d", "curl --data", "curl -X POST",
	"curl -X PUT", "wget --post", "nc ", "netcat ", "scp ", "rsync ",
	"history", ".bash_history", ".zsh_history", ".ssh/id_rsa",
	".aws/credentials", ".gnupg/", "/etc/shadow", "/etc/passwd",
	":(){ :|:& };:", "chmod +s", "mkfifo",
}

var denyPaths = []string{
	".env", ".env.local", ".env.production", "credentials.json",
	"secrets.json", "secrets.yaml", "secrets.yml", ".pem", ".key",
}

// IsSafePath checks if a path resides within the allowed directory (typically CWD).
func IsSafePath(path string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	
	// Get absolute representation
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return false
	}
	
	// Check if path escapes CWD
	rel, err := filepath.Rel(absCwd, absPath)
	if err != nil {
		return false
	}
	
	// If it starts with "..", it has escaped the CWD directory.
	if strings.HasPrefix(rel, "..") {
		return false
	}
	
	return true
}

// IsSafeCommand checks a shell command against the deny list.
// If it is blocked, it returns false and the matched deny pattern.
func IsSafeCommand(command string) (bool, string) {
	lowerCmd := strings.ToLower(command)
	for _, pattern := range denyCommands {
		if strings.Contains(lowerCmd, strings.ToLower(pattern)) {
			return false, pattern
		}
	}
	for _, pattern := range denyPaths {
		if strings.Contains(lowerCmd, strings.ToLower(pattern)) {
			return false, pattern
		}
	}
	return true, ""
}
