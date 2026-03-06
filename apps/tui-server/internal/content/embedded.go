package content

import "embed"

// embeddedContent contains the default portfolio content compiled into the Go binary.
//
//go:embed assets/*
var embeddedContent embed.FS
