package web

import "embed"

// Dist is the Vite production build (frontend outDir → backend/web/dist).
//
//go:embed all:dist
var Dist embed.FS
