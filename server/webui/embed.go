package webui

import "embed"

// FS embeds the built React UI assets at compile time.
//
//go:embed dist/*
var FS embed.FS
