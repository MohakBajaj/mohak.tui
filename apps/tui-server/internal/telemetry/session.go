package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"

	"github.com/charmbracelet/ssh"
)

// SessionInfo contains anonymized/safe session data for logging
type SessionInfo struct {
	// Anonymized identifiers (hashed for privacy)
	SessionHash string `json:"session_hash"` // Hashed session ID
	UserHash    string `json:"user_hash"`    // Hashed username
	IPHash      string `json:"ip_hash"`      // Hashed IP address

	// Non-PII technical data (safe to log)
	Terminal       string `json:"terminal"`        // Terminal type (xterm-256color, etc.)
	TerminalWidth  int    `json:"terminal_width"`  // Terminal width
	TerminalHeight int    `json:"terminal_height"` // Terminal height

	// SSH metadata (non-PII)
	ClientVersion string `json:"client_version"` // SSH client version string
	ServerVersion string `json:"server_version"` // SSH server version string

	// Connection metadata
	LocalAddr  string `json:"local_addr"`  // Server address (non-PII)
	RemotePort string `json:"remote_port"` // Only port, not full IP

	// SSH key info (anonymized)
	PublicKeyType string `json:"public_key_type,omitempty"` // Key type (ssh-rsa, ssh-ed25519, etc.)
	PublicKeyHash string `json:"public_key_hash,omitempty"` // Hashed public key fingerprint

	// Environment hints (non-PII)
	EnvTermProgram string `json:"env_term_program,omitempty"` // TERM_PROGRAM if available
	EnvShell       string `json:"env_shell,omitempty"`        // SHELL if available
	EnvLang        string `json:"env_lang,omitempty"`         // LANG/locale
	EnvColorTerm   string `json:"env_colorterm,omitempty"`    // COLORTERM if available

	// Command info
	Command    string `json:"command,omitempty"`     // Requested command
	Subsystem  string `json:"subsystem,omitempty"`   // Requested subsystem
	RawCommand string `json:"raw_command,omitempty"` // Raw command string
}

// ExtractSessionInfo extracts and anonymizes session information
func ExtractSessionInfo(s ssh.Session) SessionInfo {
	info := SessionInfo{}

	// Hash sensitive identifiers
	info.SessionHash = hashString(s.RemoteAddr().String())
	info.UserHash = hashString(s.User())

	// Extract and hash IP (keep port separate)
	if host, port, err := net.SplitHostPort(s.RemoteAddr().String()); err == nil {
		info.IPHash = hashString(host)
		info.RemotePort = port
	}

	// Local address (server-side, non-PII)
	if s.LocalAddr() != nil {
		info.LocalAddr = s.LocalAddr().String()
	}

	// PTY information
	pty, _, active := s.Pty()
	if active {
		info.Terminal = pty.Term
		info.TerminalWidth = pty.Window.Width
		info.TerminalHeight = pty.Window.Height
	}

	// SSH context information
	ctx := s.Context()

	// Client version from context (if available)
	if clientVersion, ok := ctx.Value("client_version").(string); ok {
		info.ClientVersion = clientVersion
	}

	// Server version
	if serverVersion, ok := ctx.Value("server_version").(string); ok {
		info.ServerVersion = serverVersion
	}

	// Public key info (anonymized)
	if pubKey := s.PublicKey(); pubKey != nil {
		info.PublicKeyType = pubKey.Type()
		// Hash the public key fingerprint for privacy
		info.PublicKeyHash = hashString(string(pubKey.Marshal()))
	}

	// Extract safe environment variables
	env := s.Environ()
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		// Only extract non-PII env vars
		switch key {
		case "TERM_PROGRAM":
			info.EnvTermProgram = value
		case "SHELL":
			// Only keep the shell name, not full path
			if idx := strings.LastIndex(value, "/"); idx >= 0 {
				info.EnvShell = value[idx+1:]
			} else {
				info.EnvShell = value
			}
		case "LANG", "LC_ALL":
			info.EnvLang = value
		case "COLORTERM":
			info.EnvColorTerm = value
		}
	}

	// Command information
	if cmd := s.Command(); len(cmd) > 0 {
		info.Command = cmd[0]
		info.RawCommand = strings.Join(cmd, " ")
	}
	if sub := s.Subsystem(); sub != "" {
		info.Subsystem = sub
	}

	return info
}

// ToMap converts SessionInfo to a map for logging
func (si SessionInfo) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	// Always include these
	m["session_hash"] = si.SessionHash
	m["user_hash"] = si.UserHash
	m["ip_hash"] = si.IPHash
	m["terminal"] = si.Terminal
	m["width"] = si.TerminalWidth
	m["height"] = si.TerminalHeight

	// Optional fields
	if si.ClientVersion != "" {
		m["client_version"] = si.ClientVersion
	}
	if si.PublicKeyType != "" {
		m["key_type"] = si.PublicKeyType
		m["key_hash"] = si.PublicKeyHash
	}
	if si.EnvTermProgram != "" {
		m["term_program"] = si.EnvTermProgram
	}
	if si.EnvShell != "" {
		m["shell"] = si.EnvShell
	}
	if si.EnvLang != "" {
		m["lang"] = si.EnvLang
	}
	if si.EnvColorTerm != "" {
		m["colorterm"] = si.EnvColorTerm
	}
	if si.RemotePort != "" {
		m["remote_port"] = si.RemotePort
	}

	return m
}

// hashString creates a SHA256 hash of the input, returning first 12 chars
// This provides anonymization while still allowing correlation
func hashString(input string) string {
	if input == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:12]
}

// ShortHash returns a shorter hash for display purposes
func ShortHash(input string) string {
	return hashString(input)[:8]
}
