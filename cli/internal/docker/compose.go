package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// Compose handles docker-compose operations
type Compose struct {
	projectName string
	configDir   string
	composeFile string
	envFile     string
	verbose     bool
}

// NewCompose creates a new Compose instance
func NewCompose(projectName, configDir, composeFile string) *Compose {
	return &Compose{
		projectName: projectName,
		configDir:   configDir,
		composeFile: composeFile,
		envFile:     filepath.Join(configDir, ".env"),
	}
}

// SetVerbose enables verbose output
func (c *Compose) SetVerbose(v bool) {
	c.verbose = v
}

// baseArgs returns the base docker compose arguments
func (c *Compose) baseArgs() []string {
	args := []string{
		"compose",
		"-f", c.composeFile,
		"--env-file", c.envFile,
	}
	if c.projectName != "" {
		args = append(args, "-p", c.projectName)
	}
	return args
}

// runCommand executes a docker compose command
func (c *Compose) runCommand(ctx context.Context, args []string, stream bool) error {
	fullArgs := append(c.baseArgs(), args...)

	if c.verbose {
		color.Cyan("Running: docker %s", strings.Join(fullArgs, " "))
	}

	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Dir = c.configDir
	cmd.Env = os.Environ()

	if stream {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}

	if c.verbose && len(output) > 0 {
		fmt.Println(string(output))
	}

	return nil
}

// runCommandOutput executes a command and returns output
func (c *Compose) runCommandOutput(ctx context.Context, args []string) (string, error) {
	fullArgs := append(c.baseArgs(), args...)

	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Dir = c.configDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(output))
	}

	return string(output), nil
}

// Config validates the compose configuration
func (c *Compose) Config(ctx context.Context) error {
	return c.runCommand(ctx, []string{"config", "--quiet"}, false)
}

// ConfigServices returns the list of services
func (c *Compose) ConfigServices(ctx context.Context) ([]string, error) {
	output, err := c.runCommandOutput(ctx, []string{"config", "--services"})
	if err != nil {
		return nil, err
	}

	var services []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			services = append(services, line)
		}
	}

	return services, nil
}

// Pull pulls images for all services
func (c *Compose) Pull(ctx context.Context) error {
	fmt.Println("Pulling images...")
	return c.runCommand(ctx, []string{"pull"}, true)
}

// PullService pulls images for a specific service
func (c *Compose) PullService(ctx context.Context, service string) error {
	return c.runCommand(ctx, []string{"pull", service}, true)
}

// Up starts all services
func (c *Compose) Up(ctx context.Context, detach bool, build bool) error {
	args := []string{"up"}
	if detach {
		args = append(args, "-d")
	}
	if build {
		args = append(args, "--build")
	}
	args = append(args, "--remove-orphans")

	fmt.Println("Starting services...")
	return c.runCommand(ctx, args, true)
}

// Down stops and removes all services
func (c *Compose) Down(ctx context.Context, removeVolumes bool, removeOrphans bool) error {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	if removeOrphans {
		args = append(args, "--remove-orphans")
	}

	fmt.Println("Stopping services...")
	return c.runCommand(ctx, args, true)
}

// Stop stops all services without removing them
func (c *Compose) Stop(ctx context.Context) error {
	fmt.Println("Stopping services...")
	return c.runCommand(ctx, []string{"stop"}, true)
}

// StopService stops a specific service
func (c *Compose) StopService(ctx context.Context, service string) error {
	return c.runCommand(ctx, []string{"stop", service}, true)
}

// Start starts all services
func (c *Compose) Start(ctx context.Context) error {
	fmt.Println("Starting services...")
	return c.runCommand(ctx, []string{"start"}, true)
}

// Restart restarts all services
func (c *Compose) Restart(ctx context.Context) error {
	fmt.Println("Restarting services...")
	return c.runCommand(ctx, []string{"restart"}, true)
}

// RestartService restarts a specific service
func (c *Compose) RestartService(ctx context.Context, service string) error {
	return c.runCommand(ctx, []string{"restart", service}, true)
}

// Logs streams logs for services
func (c *Compose) Logs(ctx context.Context, service string, follow bool, tail string, timestamps bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	if timestamps {
		args = append(args, "-t")
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := exec.CommandContext(ctx, "docker", append(c.baseArgs(), args...)...)
	cmd.Dir = c.configDir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Stream output
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return cmd.Wait()
}

// PS lists running containers
func (c *Compose) PS(ctx context.Context, all bool) (string, error) {
	args := []string{"ps"}
	if all {
		args = append(args, "-a")
	}
	return c.runCommandOutput(ctx, args)
}

// Exec executes a command in a running container
func (c *Compose) Exec(ctx context.Context, service string, command []string, interactive bool) error {
	args := []string{"exec"}
	if interactive {
		args = append(args, "-it")
	}
	args = append(args, service)
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "docker", append(c.baseArgs(), args...)...)
	cmd.Dir = c.configDir
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Run runs a one-off command in a new container
func (c *Compose) Run(ctx context.Context, service string, command []string, rm bool) error {
	args := []string{"run"}
	if rm {
		args = append(args, "--rm")
	}
	args = append(args, service)
	args = append(args, command...)

	return c.runCommand(ctx, args, true)
}

// GetContainerID returns the container ID for a service
func (c *Compose) GetContainerID(ctx context.Context, service string) (string, error) {
	output, err := c.runCommandOutput(ctx, []string{"ps", "-q", service})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// IsRunning checks if a service is running
func (c *Compose) IsRunning(ctx context.Context, service string) (bool, error) {
	id, err := c.GetContainerID(ctx, service)
	if err != nil {
		return false, err
	}
	return id != "", nil
}

// CheckComposeInstalled verifies docker compose is available
func CheckComposeInstalled() error {
	cmd := exec.Command("docker", "compose", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose is not installed or not accessible: %w\n%s", err, string(output))
	}
	return nil
}
