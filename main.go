package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// getContainerFromID retrieves container information by container ID
func getContainerFromID(cli *client.Client, containerID string) (*types.ContainerJSON, error) {
	ctx := context.Background()

	// Get detailed container information
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %v", containerID, err)
	}

	return &containerInfo, nil
}

// getStateJSON fetches the state.json file from the Docker runtime directory
func getStateJSON(containerID string) ([]byte, error) {
	// Construct the path to the state.json file
	statePath := filepath.Join("/var/run/docker/runtime-runc/moby", containerID, "state.json")

	// Check if the file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("state.json file not found at %s", statePath)
	}

	// Read the state.json file
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state.json file: %v", err)
	}

	return stateData, nil
}

// createAndStartContainer creates a new container from an image and starts it
// If sourceContainerID is provided, the new container will share namespaces with the source
func createAndStartContainer(cli *client.Client, imageName string, containerName string, sourceContainerID string) (string, error) {
	ctx := context.Background()

	hostConfig := &container.HostConfig{
		AutoRemove: false,
	}

	// If source container ID is provided, share namespaces with it
	if sourceContainerID != "" {
		hostConfig.PidMode = container.PidMode("container:" + sourceContainerID)
		hostConfig.NetworkMode = container.NetworkMode("container:" + sourceContainerID)
		hostConfig.IpcMode = container.IpcMode("container:" + sourceContainerID)
	}

	config := &container.Config{
		Image:        imageName,
		Cmd:          []string{"/bin/sh", "-c", "sleep infinity"}, // this is to Keep container running
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		OpenStdin:    false,
		StdinOnce:    false,
		Tty:          false, // Set to false for true detached mode
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	containerID := resp.ID
	fmt.Printf("Created container: %s (ID: %s)\n", containerName, containerID)

	// Start the container
	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to start container %s: %v", containerID, err)
	}

	fmt.Printf("Successfully started container: %s\n", containerID)
	return containerID, nil
}

func copyStateJSON(cli *client.Client, source string, destination string, stateJSON []byte) error {
	ctx := context.Background()

	destStatePath := filepath.Join("/var/run/docker/runtime-runc/moby", destination, "state.json")

	err := os.WriteFile(destStatePath, stateJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state.json to destination container %s: %v", destination, err)
	}
	fmt.Printf("Copied state.json from %s to %s\n", source, destination)

	// Start the destination container with the new state
	err = cli.ContainerStart(ctx, destination, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start destination container %s: %v", destination, err)
	}
	fmt.Printf("Started destination container with copied state: %s\n", destination)

	return nil
}
func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.45"))
	if err != nil {
		panic(err)
	}

	// Example usage: get container by ID
	var containerID string
	fmt.Print("Enter container ID: ")
	fmt.Scan(&containerID)

	containerInfo, err := getContainerFromID(cli, containerID)
	if err != nil {
		log.Fatalf("Error getting container: %v", err)
	}

	// Print container information
	fmt.Printf("Container ID: %s\n", containerInfo.ID)
	fmt.Printf("Container Name: %s\n", containerInfo.Name)
	fmt.Printf("Container State: %s\n", containerInfo.State.Status)
	fmt.Printf("Container Image: %s\n", containerInfo.Config.Image)

	// Fetch the state.json file
	stateData, err := getStateJSON(containerInfo.ID)
	if err != nil {
		log.Printf("Warning: Could not fetch state.json: %v", err)
	} else {
		fmt.Printf("State.json file size: %d bytes\n", len(stateData))
	}

	// Create and start Alpine container named "copy" in background, sharing namespaces with source
	copyContainerID, err := createAndStartContainer(cli, "alpine", "cloned-cont", containerInfo.ID)
	if err != nil {
		log.Printf("Failed to create copy container: %v", err)
	} else {
		fmt.Printf("Successfully created and started 'copy' container: %s\n", copyContainerID)

		// Copy state.json from source container to destination container
		if len(stateData) > 0 {
			err = copyStateJSON(cli, containerInfo.ID, copyContainerID, stateData)
			if err != nil {
				log.Printf("Failed to copy state.json: %v", err)
			} else {
				fmt.Println("Successfully copied state.json between containers!")
			}
		}
	}
}
