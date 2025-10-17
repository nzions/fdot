package netssh

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/nzions/fdot/pkg/fdh/credmgr"
	"github.com/nzions/fdot/pkg/fdh/netmodel"
	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client configured for network devices
type Client struct {
	config *ssh.ClientConfig
	conn   *ssh.Client
	host   string
	port   int
	cache  *netmodel.CommandCache
}

// Config holds configuration for creating a network SSH client
type Config struct {
	Host        string
	Port        int
	Credentials credmgr.UserCred
	Timeout     time.Duration
	CacheConfig *netmodel.CacheConfig // Optional cache configuration
}

// NewClient creates a new SSH client configured for network devices
func NewClient(ctx context.Context, cfg Config) *Client {
	if cfg.Port == 0 {
		cfg.Port = 22 // Default SSH port
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second // Default timeout
	}
	if cfg.CacheConfig == nil {
		cfg.CacheConfig = netmodel.DefaultCacheConfig()
	}

	return &Client{
		config: &ssh.ClientConfig{
			User: cfg.Credentials.Username(),
			Auth: []ssh.AuthMethod{
				ssh.Password(cfg.Credentials.Password()),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For network devices, typically don't validate host keys
			Timeout:         cfg.Timeout,
		},
		host:  cfg.Host,
		port:  cfg.Port,
		cache: netmodel.NewCommandCache(cfg.CacheConfig),
	}
}

// ExecuteOption is a functional option for configuring command execution
type ExecuteOption func(*executeOptions)

// executeOptions holds the configuration for command execution
type executeOptions struct {
	noCache bool
	timeout time.Duration
}

// OptNoCache disables caching for this command execution
func OptNoCache() ExecuteOption {
	return func(opts *executeOptions) {
		opts.noCache = true
	}
}

// OptTimeout sets a custom timeout for this command execution
func OptTimeout(timeout time.Duration) ExecuteOption {
	return func(opts *executeOptions) {
		opts.timeout = timeout
	}
}

// Connect establishes the SSH connection
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", addr, err)
	}
	c.conn = conn
	return nil
}

// ExecuteCommand executes a command on the remote device and returns the output
// Supports functional options for configuration (OptNoCache, OptTimeout, etc.)
func (c *Client) ExecuteCommand(cmd string, opts ...ExecuteOption) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected - call Connect() first")
	}

	// Parse options
	execOpts := &executeOptions{
		noCache: false,
		timeout: time.Duration(time.Second * 30),
	}
	for _, opt := range opts {
		opt(execOpts)
	}

	// Check cache first (unless disabled)
	if !execOpts.noCache {
		if cachedOutput, found := c.cache.GetCachedOutput(c.host, cmd); found {
			return cachedOutput, nil
		}
	}

	// Execute the command
	output, err := c.executeCommandInternal(cmd, execOpts)
	if err != nil {
		return "", err
	}

	// Save to cache (unless disabled)
	if !execOpts.noCache {
		_ = c.cache.SaveOutput(c.host, cmd, output)
	}

	return output, nil
}

// executeCommandInternal performs the actual SSH command execution
func (c *Client) executeCommandInternal(cmd string, opts *executeOptions) (string, error) {
	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set terminal modes for network devices
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal for interactive commands
	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		return "", fmt.Errorf("request for pseudo terminal failed: %w", err)
	}

	// Get pipes for reading output
	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Create channels for reading output with timeout support
	type result struct {
		output string
		err    error
	}
	resultChan := make(chan result, 1)

	// Read output in goroutine to support timeout
	go func() {
		outputBytes, err := io.ReadAll(stdout)
		if err != nil {
			resultChan <- result{"", fmt.Errorf("failed to read stdout: %w", err)}
			return
		}

		errBytes, err := io.ReadAll(stderr)
		if err != nil {
			resultChan <- result{"", fmt.Errorf("failed to read stderr: %w", err)}
			return
		}

		// Wait for command to complete
		waitErr := session.Wait()
		if waitErr != nil {
			// Some network devices return non-zero exit codes even on success
			// Don't fail if we got output
			if len(outputBytes) == 0 {
				resultChan <- result{"", fmt.Errorf("command failed: %w (stderr: %s)", waitErr, string(errBytes))}
				return
			}
		}

		output := string(outputBytes)
		if len(errBytes) > 0 {
			output += "\n" + string(errBytes)
		}

		resultChan <- result{output, nil}
	}()

	// Wait for result with timeout
	select {
	case res := <-resultChan:
		return res.output, res.err
	case <-time.After(opts.timeout):
		return "", fmt.Errorf("command execution timed out after %v", opts.timeout)
	}
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
