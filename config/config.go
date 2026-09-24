package config

import "time"

const (
	CLIENT_PREFIX  = "Vk"
	CLIENT_VERSION = "0.1.0"

	TRACKER_TIMEOUT_DURATION = 6 * time.Second

	PEER_HANDSHAKE_TIMEOUT_DURATION = 6 * time.Second
	PEER_KEEPALIVE_TIMEOUT_DURATION = 2 * time.Minute
)
