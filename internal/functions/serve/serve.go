package serve

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/functions/serve/docker"
	"github.com/protomoks/pmok/internal/utils"
	"github.com/protomoks/pmok/internal/utils/constants"
)

//go:embed templates/local-main.ts
var mainFunc string

func Run(ctx context.Context, cm docker.ContainerManager) error {
	conf := config.GetConfig()
	if conf == nil {
		return utils.ConfigNotFound()
	}
	// remove the container
	_ = cm.KillAndRemoveContainer(ctx, constants.FunctionsServerContainer, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	// pull the image
	if err := cm.PullImage(ctx, constants.DenoImage, os.Stderr); err != nil {
		return err
	}
	// fnConfig, err := conf.Manifest.Functions.ToJSON()
	// if err != nil {
	// 	return err
	// }
	env := []string{
		//fmt.Sprintf("PROTOMOK_FUNCTION_CONFIG=%s", string(fnConfig)),
		fmt.Sprintf("PROTOMOK_CONFIG_ENCODING=%s", string(conf.Manifest.Encoding())),
	}

	// cmd := []string{
	// 	"edge-runtime",
	// 	"start",
	// 	"--main-service=/root",
	// 	fmt.Sprintf("--port=%d", 8082),
	// 	"--verbose",
	// 	fmt.Sprintf("--policy=%s", "oneshot"),
	// }
	cmd := []string{
		"deno",
		"run",
		"--allow-net",
		"--allow-read",
		"--allow-env",
		"/root/index.ts",
	}
	cmdStr := strings.Join(cmd, " ")
	entryPoint := []string{"sh", "-c", `cat <<'EOF' > /root/index.ts && ` + cmdStr + `
	` + mainFunc + `
EOF
`}

	// create the container TODO: Delete existing container if it exists
	id, err := cm.CreateContainer(ctx,
		&container.Config{
			Env: env,
			//Image: constants.EdgeRuntimeImage,
			Image:        constants.DenoImage,
			Entrypoint:   entryPoint,
			ExposedPorts: nat.PortSet{nat.Port(fmt.Sprintf("%d/tcp", 8000)): struct{}{}},
			WorkingDir:   utils.Slashify(conf.GetProjectDir()),
		},
		&container.HostConfig{
			Binds: createBinds(conf),
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%d/tcp", 8000)): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "8000",
					},
				},
			},
		},
		constants.FunctionsServerContainer,
	)

	if err != nil {
		return err
	}

	fmt.Printf("Container with id %s created\n", id)

	err = cm.StartContainer(ctx, id, container.StartOptions{})
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func createBinds(conf *config.Config) []string {
	binds := []string{
		constants.FunctionsServerContainer + ":" + "/root/.cache/deno:rw",
		filepath.Join(conf.GetProjectDir(), config.ProtomokDir) + ":" + utils.Slashify(filepath.Join(conf.GetProjectDir(), config.ProtomokDir)),
	}
	return binds
}
