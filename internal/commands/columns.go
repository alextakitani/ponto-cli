package commands

import "github.com/alextakitani/ponto-cli/internal/render"

// Column definitions for styled/markdown table rendering.
var (
	authProfileColumns = render.Columns{
		{Header: "Profile", Field: "profile"},
		{Header: "Active", Field: "active"},
		{Header: "Base URL", Field: "base_url"},
	}
	entryColumns = render.Columns{
		{Header: "ID", Field: "id"},
		{Header: "Description", Field: "description"},
		{Header: "Project", Field: "project_id"},
		{Header: "Started", Field: "started_at"},
		{Header: "Duration", Field: "duration"},
	}
	clientColumns = render.Columns{
		{Header: "ID", Field: "id"},
		{Header: "Name", Field: "name"},
		{Header: "Rate", Field: "rate"},
		{Header: "Archived", Field: "archived_at"},
	}
	projectColumns = render.Columns{
		{Header: "ID", Field: "id"},
		{Header: "Name", Field: "name"},
		{Header: "Client", Field: "client_id"},
		{Header: "Rate", Field: "effective_rate"},
		{Header: "Default", Field: "default"},
	}
	taskColumns = render.Columns{
		{Header: "ID", Field: "id"},
		{Header: "Name", Field: "name"},
		{Header: "Project", Field: "project_id"},
	}
	tagColumns = render.Columns{
		{Header: "ID", Field: "id"},
		{Header: "Name", Field: "name"},
		{Header: "Archived", Field: "archived_at"},
	}
)
