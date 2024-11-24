package redispkg

// Config holds redis config
type Config struct {
	Address            string
	Password           string
	TLSEnabled         bool
	ClusterModeEnabled bool
}
