// Package mcphost provides MCP (Model Context Protocol) server management functionality.
// It handles configuration loading, server connection management, and tool discovery
// for various MCP server implementations including stdio and SSE transports.
package mcphost

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Timeout configuration constants
const (
	// DefaultMCPTimeoutSeconds is the default timeout for MCP operations
	DefaultMCPTimeoutSeconds = 30
	// MinMCPTimeoutSeconds is the minimum allowed timeout for MCP operations
	MinMCPTimeoutSeconds = 5
)

// Error messages
const (
	errMsgSettingsNil           = "settings cannot be nil"
	errMsgTimeoutTooSmall       = "server %s: timeout must be at least %d seconds"
	errMsgURLRequired           = "server %s: URL is required for SSE transport"
	errMsgCommandRequired       = "server %s: command is required for stdio transport"
	errMsgUnsupportedTransport  = "server %s: unsupported transport type: %s"
	errMsgFailedToParseSettings = "failed to parse settings: %w"
	errMsgInvalidSettings       = "invalid settings: %w"
	errMsgFailedToReadFile      = "failed to read settings file: %w"
)

// MCPSettings represents the main configuration structure for MCP servers.
// It contains a map of server configurations indexed by server name.
type MCPSettings struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents configuration for a single MCP server.
// It supports both SSE and stdio transport types with their respective settings.
type ServerConfig struct {
	TransportType string        `json:"transportType,omitempty"` // "sse" or "stdio"
	AutoApprove   []string      `json:"autoApprove,omitempty"`   // List of auto-approved operations
	Disabled      bool          `json:"disabled,omitempty"`      // Whether the server is disabled
	Timeout       time.Duration `json:"timeout,omitempty"`       // Operation timeout

	// SSE specific configuration
	URL string `json:"url,omitempty"` // Server URL for SSE transport

	// Stdio specific configuration
	Command string            `json:"command"`       // Command to execute for stdio transport
	Args    []string          `json:"args"`          // Command arguments
	Env     map[string]string `json:"env,omitempty"` // Environment variables
}

// GetTimeoutDuration returns the timeout duration for the server.
// If no timeout is configured, it returns the default timeout.
func (c *ServerConfig) GetTimeoutDuration() time.Duration {
	if c.Timeout == 0 {
		return time.Duration(DefaultMCPTimeoutSeconds) * time.Second
	}
	return c.Timeout
}

// IsSSETransport returns true if the server uses SSE transport.
func (c *ServerConfig) IsSSETransport() bool {
	return c.TransportType == TransportTypeSSE
}

// IsStdioTransport returns true if the server uses stdio transport.
func (c *ServerConfig) IsStdioTransport() bool {
	return c.TransportType == TransportTypeStdio || c.TransportType == ""
}

// LoadSettingsFromString loads MCP settings from a JSON configuration string.
// It parses the JSON and validates the resulting configuration.
//
// Parameters:
//   - data: JSON configuration string
//
// Returns:
//   - *MCPSettings: Parsed and validated settings
//   - error: Error if parsing or validation fails
func LoadSettingsFromString(data string) (*MCPSettings, error) {
	if strings.TrimSpace(data) == "" {
		return &MCPSettings{MCPServers: make(map[string]ServerConfig)}, nil
	}

	var settings MCPSettings
	if err := json.Unmarshal([]byte(data), &settings); err != nil {
		return nil, fmt.Errorf(errMsgFailedToParseSettings, err)
	}

	if err := validateSettings(&settings); err != nil {
		return nil, fmt.Errorf(errMsgInvalidSettings, err)
	}

	return &settings, nil
}

// LoadSettings loads MCP settings from a configuration file.
// It reads the file and delegates to LoadSettingsFromString for parsing.
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *MCPSettings: Parsed and validated settings
//   - error: Error if file reading, parsing, or validation fails
func LoadSettings(path string) (*MCPSettings, error) {
	log.Printf("Loading MCP settings from: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(errMsgFailedToReadFile, err)
	}

	return LoadSettingsFromString(string(data))
}

// validateSettings validates the MCP settings configuration.
// It checks for required fields and validates timeout values.
func validateSettings(settings *MCPSettings) error {
	if settings == nil {
		return fmt.Errorf(errMsgSettingsNil)
	}

	// Initialize MCPServers map if nil
	if settings.MCPServers == nil {
		settings.MCPServers = make(map[string]ServerConfig)
	}

	for name, server := range settings.MCPServers {
		if err := validateServerConfig(name, server); err != nil {
			return err
		}
	}

	return nil
}

// validateServerConfig validates a single server configuration.
func validateServerConfig(name string, server ServerConfig) error {
	// Validate timeout
	if server.Timeout > 0 && server.Timeout < time.Duration(MinMCPTimeoutSeconds)*time.Second {
		return fmt.Errorf(errMsgTimeoutTooSmall, name, MinMCPTimeoutSeconds)
	}

	// Validate transport-specific requirements
	switch server.TransportType {
	case TransportTypeSSE:
		if strings.TrimSpace(server.URL) == "" {
			return fmt.Errorf(errMsgURLRequired, name)
		}
	case TransportTypeStdio, "":
		if strings.TrimSpace(server.Command) == "" {
			return fmt.Errorf(errMsgCommandRequired, name)
		}
	default:
		return fmt.Errorf(errMsgUnsupportedTransport, name, server.TransportType)
	}

	return nil
}
