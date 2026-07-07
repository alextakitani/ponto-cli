package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type quickStartResponse struct {
	Version  string                 `json:"version"`
	Auth     quickStartAuthInfo     `json:"auth"`
	Context  quickStartContextInfo  `json:"context"`
	Commands quickStartCommandsInfo `json:"commands"`
}

type quickStartAuthInfo struct {
	Status  string `json:"status"`
	Profile string `json:"profile,omitempty"`
}

type quickStartContextInfo struct {
	APIURL string `json:"api_url,omitempty"`
}

type quickStartCommandsInfo struct {
	QuickStart []string `json:"quick_start"`
	Common     []string `json:"common"`
}

func runRootDefault(cmd *cobra.Command, args []string) error {
	if isHumanOutput() {
		return cmd.Help()
	}

	auth := quickStartAuthInfo{Status: "unauthenticated"}
	if cfgProfile != "" {
		auth.Profile = cfgProfile
	}
	if cfg != nil {
		if cfg.Profile != "" {
			auth.Profile = cfg.Profile
		}
		if cfg.Token != "" {
			auth.Status = "authenticated"
		}
	}

	context := quickStartContextInfo{}
	if cfg != nil && cfg.APIURL != "" {
		context.APIURL = cfg.APIURL
	}

	resp := quickStartResponse{
		Version: currentVersion(),
		Auth:    auth,
		Context: context,
		Commands: quickStartCommandsInfo{
			QuickStart: []string{"ponto setup", "ponto doctor", "ponto config show", "ponto commands"},
			Common:     []string{"ponto auth status", "ponto auth list", "ponto config explain", "ponto version"},
		},
	}

	summary := fmt.Sprintf("ponto %s - not logged in", currentVersion())
	if auth.Status == "authenticated" {
		summary = fmt.Sprintf("ponto %s - logged in", currentVersion())
	}
	if auth.Profile != "" {
		summary += fmt.Sprintf(" (profile: %s)", auth.Profile)
	}

	breadcrumbs := []Breadcrumb{
		breadcrumb("doctor", "ponto doctor", "Check CLI health"),
		breadcrumb("config", "ponto config show", "Show the effective config"),
		breadcrumb("commands", "ponto commands", "List available commands"),
	}
	if auth.Status == "unauthenticated" {
		breadcrumbs = append(breadcrumbs, breadcrumb("authenticate", authLoginHint(), "Authenticate"))
	}

	printSuccessWithBreadcrumbs(resp, summary, breadcrumbs)
	return nil
}

func authLoginHint() string {
	parts := []string{"ponto", "auth", "login", "<token>"}
	if strings.TrimSpace(cfgProfile) != "" {
		parts = append(parts, "--profile", cfgProfile)
	}
	return strings.Join(parts, " ")
}
