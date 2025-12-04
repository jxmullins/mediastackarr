package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// Client wraps the Docker SDK client
type Client struct {
	cli         *client.Client
	projectName string
}

// ContainerInfo holds information about a container
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	State   string
	Status  string
	Health  string
	Ports   []string
	Created int64
}

// NewClient creates a new Docker client
func NewClient(projectName string) (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		cli:         cli,
		projectName: projectName,
	}, nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.cli.Close()
}

// Ping tests the Docker connection
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.cli.Ping(ctx)
	return err
}

// ListContainers returns all containers for the project
func (c *Client) ListContainers(ctx context.Context, all bool) ([]ContainerInfo, error) {
	filterArgs := filters.NewArgs()
	if c.projectName != "" {
		filterArgs.Add("label", fmt.Sprintf("com.docker.compose.project=%s", c.projectName))
	}

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     all,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, cont := range containers {
		info := ContainerInfo{
			ID:      cont.ID[:12],
			Name:    strings.TrimPrefix(cont.Names[0], "/"),
			Image:   cont.Image,
			State:   cont.State,
			Status:  cont.Status,
			Created: cont.Created,
		}

		// Get health status if available
		if cont.State == "running" {
			inspect, err := c.cli.ContainerInspect(ctx, cont.ID)
			if err == nil && inspect.State.Health != nil {
				info.Health = inspect.State.Health.Status
			}
		}

		// Format ports
		for _, p := range cont.Ports {
			if p.PublicPort > 0 {
				info.Ports = append(info.Ports, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// StopContainer stops a container by ID or name
func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30 // seconds
	return c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

// RemoveContainer removes a container by ID or name
func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         force,
		RemoveVolumes: false,
	})
}

// StopAllProjectContainers stops all containers belonging to the project
func (c *Client) StopAllProjectContainers(ctx context.Context) error {
	containers, err := c.ListContainers(ctx, false) // only running
	if err != nil {
		return err
	}

	for _, cont := range containers {
		if err := c.StopContainer(ctx, cont.ID); err != nil {
			return fmt.Errorf("failed to stop container %s: %w", cont.Name, err)
		}
	}

	return nil
}

// RemoveAllProjectContainers removes all containers belonging to the project
func (c *Client) RemoveAllProjectContainers(ctx context.Context) error {
	containers, err := c.ListContainers(ctx, true) // all containers
	if err != nil {
		return err
	}

	for _, cont := range containers {
		if err := c.RemoveContainer(ctx, cont.ID, true); err != nil {
			return fmt.Errorf("failed to remove container %s: %w", cont.Name, err)
		}
	}

	return nil
}

// GetContainerLogs returns logs from a container
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, follow bool, tail string) (io.ReadCloser, error) {
	return c.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	})
}

// ContainerExec executes a command in a container
func (c *Client) ContainerExec(ctx context.Context, containerID string, cmd []string) (string, error) {
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execID, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := c.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	return string(output), nil
}

// PruneContainers removes all stopped containers
func (c *Client) PruneContainers(ctx context.Context) error {
	_, err := c.cli.ContainersPrune(ctx, filters.Args{})
	return err
}

// PruneVolumes removes all unused volumes
func (c *Client) PruneVolumes(ctx context.Context) error {
	_, err := c.cli.VolumesPrune(ctx, filters.Args{})
	return err
}

// PruneNetworks removes all unused networks
func (c *Client) PruneNetworks(ctx context.Context) error {
	_, err := c.cli.NetworksPrune(ctx, filters.Args{})
	return err
}

// PruneImages removes all unused images
func (c *Client) PruneImages(ctx context.Context) error {
	_, err := c.cli.ImagesPrune(ctx, filters.NewArgs(filters.Arg("dangling", "false")))
	return err
}

// PullImage pulls a Docker image
func (c *Client) PullImage(ctx context.Context, imageName string) error {
	out, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	// Discard output (could be used for progress in future)
	_, err = io.Copy(io.Discard, out)
	return err
}

// FindContainer finds a container by service name
func (c *Client) FindContainer(ctx context.Context, serviceName string) (*ContainerInfo, error) {
	containers, err := c.ListContainers(ctx, true)
	if err != nil {
		return nil, err
	}

	for _, cont := range containers {
		// Check if container name contains service name
		if strings.Contains(cont.Name, serviceName) {
			return &cont, nil
		}
	}

	return nil, fmt.Errorf("container not found for service: %s", serviceName)
}

// ReadFileFromContainer reads a file from inside a container
func (c *Client) ReadFileFromContainer(ctx context.Context, containerID, filePath string) ([]byte, error) {
	reader, _, err := c.cli.CopyFromContainer(ctx, containerID, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file from container: %w", err)
	}
	defer reader.Close()

	// The response is a tar archive, we need to extract it
	return io.ReadAll(reader)
}

// CheckDockerRunning verifies Docker daemon is accessible
func CheckDockerRunning() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	_, err = cli.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("Docker daemon is not running: %w", err)
	}

	return nil
}

// IsRunningInDocker returns true if the current process is running inside Docker
func IsRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
