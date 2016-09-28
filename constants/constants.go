package constants

import (
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
)
