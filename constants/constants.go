package constants

const (
	CTX_DB            = "db"
	CTX_DOCKER_CLIENT = "docker_client"

	BUILD_TIMEOUT   = 5
	BUILD_CPU_SHARE = 102      // cpu share relative to the default (1024) for build containers
	BUILD_MEMORY    = 52428800 // memory in bytes for build containers (50mb)
)
