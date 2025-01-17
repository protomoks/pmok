package docker

import (
	"context"
	"io"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
)

type ContainerManager interface {
	PullImage(ctx context.Context, img string, w io.Writer) error
	CreateContainer(ctx context.Context, config *container.Config, hostconfig *container.HostConfig, name string) (string, error)
	StartContainer(ctx context.Context, id string, options container.StartOptions) error
	KillAndRemoveContainer(ctx context.Context, id string, options container.RemoveOptions) error
}

type DockerClient interface {
	client.ImageAPIClient
	client.ContainerAPIClient
}

type containermanager struct {
	cli DockerClient
}

func NewContainerManager() (ContainerManager, error) {
	cli, err := initDockerClient()
	if err != nil {
		return nil, errors.Errorf("unable to create container manager %s", err)
	}
	return &containermanager{cli: cli}, nil
}

func initDockerClient() (DockerClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func (c *containermanager) PullImage(ctx context.Context, img string, w io.Writer) error {
	out, err := c.cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		return errors.Wrap(err, "image pull")
	}

	defer func() {
		if err := out.Close(); err != nil {
			// log error here somehow
		}
	}()
	if err := jsonmessage.DisplayJSONMessagesToStream(out, streams.NewOut(w), nil); err != nil {
		return errors.Errorf("error streaming image pull output %s", err)
	}
	return nil
}

func (c *containermanager) CreateContainer(ctx context.Context, config *container.Config, hostconfig *container.HostConfig, name string) (string, error) {
	res, err := c.cli.ContainerCreate(ctx, config, hostconfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	return res.ID, err
}

func (c *containermanager) StartContainer(ctx context.Context, id string, options container.StartOptions) error {
	return c.cli.ContainerStart(ctx, id, options)
}

func (c *containermanager) KillAndRemoveContainer(ctx context.Context, id string, options container.RemoveOptions) error {
	return c.cli.ContainerRemove(ctx, id, options)
}
