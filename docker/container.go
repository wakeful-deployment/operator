package docker

import (
	"fmt"
	"os"
	"strings"
)

type Container interface {
	Name() string
	Image() string
	Ports() []string
	Env() map[string]string
	Restart() string
	Tags() []string
}

func portsArgs(ports []string) []string {
	if len(ports) == 0 {
		return []string{"-P"}
	}

	var args []string
	for _, port := range ports {
		args = append(args, "-p", port)
	}
	return args
}

func envArgs(vars map[string]string) []string {
	var args []string

	for key, value := range vars {
		if strings.HasPrefix(value, "$") && strings.ToUpper(value) == value {
			value, _ = os.LookupEnv(value[1:len(value)])
		}
		str := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-e", str)
	}

	return args
}

func restartArg(setting string) string {
	if setting == "" {
		return "--restart=always"
	}

	return fmt.Sprintf("--restart=%s", setting)
}

func RunArgs(c Container) []string {
	args := []string{"run", "-d", "--name", c.Name()}
	args = append(args, portsArgs(c.Ports())...)
	args = append(args, envArgs(c.Env())...)
	args = append(args, restartArg(c.Restart()))
	args = append(args, c.Image())

	var cleaned []string
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			cleaned = append(cleaned, arg)
		}
	}

	return cleaned
}
