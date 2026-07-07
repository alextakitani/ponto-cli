package commands

import "github.com/alextakitani/ponto-cli/internal/render"

// Column definitions for styled/markdown table rendering.
var (
	authProfileColumns = render.Columns{
		{Header: "Profile", Field: "profile"},
		{Header: "Active", Field: "active"},
		{Header: "Base URL", Field: "base_url"},
	}
)
