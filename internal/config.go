package internal

type Config struct {
	// key value store dir
	KVDir string `toml:"kv_dir"`

	// hearbeat timeout
	HeartbeatTimeout uint `toml:"heartbeat_timeout"`
}
