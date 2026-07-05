package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	defaultGoImage     = "golang:1.21-alpine"
	defaultPythonImage = "python:3.12-alpine"
	defaultTimeout     = 60 * time.Second
	memoryLimit        = 512 * 1024 * 1024
	nanoCPUs           = 1_000_000_000
)

type Result struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type DockerRunner struct {
	specs   map[Language]Spec
	timeout time.Duration
	cli     *client.Client
}

func NewDockerRunner(goImage, pythonImage string, timeout time.Duration) (*DockerRunner, error) {
	if goImage == "" {
		goImage = defaultGoImage
	}
	if pythonImage == "" {
		pythonImage = defaultPythonImage
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerRunner{
		specs:   buildSpecs(goImage, pythonImage),
		timeout: timeout,
		cli:     cli,
	}, nil
}

func (r *DockerRunner) Run(ctx context.Context, lang Language, code string) Result {
	spec, ok := r.specs[lang]
	if !ok {
		return Result{Error: fmt.Sprintf("unsupported language: %s", lang)}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	deadline, ok := ctx.Deadline()
	timeout := r.timeout
	if ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	containerID, err := r.createContainer(runCtx, spec)
	if err != nil {
		return Result{Error: err.Error()}
	}
	defer r.removeContainer(context.Background(), containerID)

	if err := r.copyCode(runCtx, containerID, spec, code); err != nil {
		return Result{Error: "failed to copy code into container: " + err.Error()}
	}

	if err := r.cli.ContainerStart(runCtx, containerID, container.StartOptions{}); err != nil {
		return Result{Error: "failed to start container: " + err.Error()}
	}

	logsReader, err := r.cli.ContainerLogs(runCtx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return Result{Error: "failed to read logs: " + err.Error()}
	}
	defer logsReader.Close()

	var outBuf, errBuf bytes.Buffer
	logDone := make(chan struct{})
	go func() {
		defer close(logDone)
		_, _ = stdcopy.StdCopy(&outBuf, &errBuf, logsReader)
	}()

	waitCh, errCh := r.cli.ContainerWait(runCtx, containerID, container.WaitConditionNotRunning)

	var waitResult container.WaitResponse
	select {
	case waitResult = <-waitCh:
	case err := <-errCh:
		return Result{Error: "container wait failed: " + err.Error()}
	case <-runCtx.Done():
		_ = r.cli.ContainerKill(context.Background(), containerID, "SIGKILL")
		<-logDone
		combined := combineOutput(outBuf.String(), errBuf.String())
		return Result{
			Output: combined,
			Error:  "execution timeout",
		}
	}

	<-logDone

	stdout := strings.TrimSpace(outBuf.String())
	stderr := strings.TrimSpace(errBuf.String())
	combined := combineOutput(stdout, stderr)

	if waitResult.StatusCode != 0 {
		errMsg := stderr
		if errMsg == "" {
			errMsg = "program exited with non-zero status"
		}
		return Result{
			Output: combined,
			Error:  errMsg,
		}
	}

	if stderr != "" && stdout != "" {
		return Result{Output: combined}
	}
	if stderr != "" {
		return Result{Output: stderr}
	}
	return Result{Output: stdout}
}

func (r *DockerRunner) createContainer(ctx context.Context, spec Spec) (string, error) {
	resp, err := r.cli.ContainerCreate(ctx, &container.Config{
		Image:      spec.Image,
		Cmd:        spec.Cmd,
		WorkingDir: "/workspace",
	}, &container.HostConfig{
		NetworkMode: "none",
		Resources: container.Resources{
			Memory:   memoryLimit,
			NanoCPUs: nanoCPUs,
		},
	}, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (r *DockerRunner) copyCode(ctx context.Context, containerID string, spec Spec, code string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name: spec.Filename,
		Mode: 0o644,
		Size: int64(len(code)),
	}); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(code)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return r.cli.CopyToContainer(ctx, containerID, "/workspace", &buf, types.CopyToContainerOptions{})
}

func (r *DockerRunner) removeContainer(ctx context.Context, id string) {
	_ = r.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

func combineOutput(stdout, stderr string) string {
	stdout = strings.TrimSpace(stdout)
	stderr = strings.TrimSpace(stderr)
	switch {
	case stdout != "" && stderr != "":
		return stdout + "\n" + stderr
	case stderr != "":
		return stderr
	default:
		return stdout
	}
}

// Close releases the Docker client.
func (r *DockerRunner) Close() error {
	if r.cli == nil {
		return nil
	}
	return r.cli.Close()
}
