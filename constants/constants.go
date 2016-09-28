package constants

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/strslice"
)

const (
	CTX_DB            = "db"
	CTX_DOCKER_CLIENT = "docker_client"

	BUILD_TIMEOUT   = 5
	BUILD_CPU_SHARE = 102      // cpu share relative to the default (1024) for build containers
	BUILD_MEMORY    = 52428800 // memory in bytes for build containers (50mb)

	HASH_LENGTH = 8

	GITHUB_CLIENT_ID     = "47ecbefcf49c1c3ce7d4"
	GITHUB_CLIENT_SECRET = "02c1e09bca7d4f270f93852aecc5a4315f3c1822"

	JWT_SECRET = "859742d8bfc747a4aa729291d710aace"
)

type defaultContainerConfig struct {
	image   string
	command string
}

func (d defaultContainerConfig) Command(args ...interface{}) strslice.StrSlice {
	return strslice.StrSlice{"sh", "-c", fmt.Sprintf(d.command, args...)}
}

func (d defaultContainerConfig) Image() string {
	return d.image
}

type DefaultContainerConfig interface {
	Image() string
	Command(...interface{}) strslice.StrSlice
}

var cppConf = defaultContainerConfig{
	image:   "frolvlad/alpine-gcc:latest",
	command: "g++ %s && ./a.out",
}

// some var types cannot be const :(
var (
	FILETYPE_CONFIGS = map[string]DefaultContainerConfig{
		".cpp": cppConf,
		".cxx": cppConf,
		".java": defaultContainerConfig{
			image:   "openjdk:8-alpine",
			command: "javac %s && java Solution",
		},
	}

	UNRECOGNIZED_HASH = errors.New("The provided hash could not be found")
)
