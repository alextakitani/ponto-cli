package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alextakitani/ponto-cli/internal/config"
	"github.com/alextakitani/ponto-cli/internal/errors"
	"github.com/basecamp/cli/output"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long:  "Configure Ponto CLI with your API URL and API token.",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	if cfgJQ != "" {
		return errors.ErrJQNotSupported("the setup command")
	}
	if IsMachineOutput() {
		return output.ErrUsageHint("setup requires an interactive terminal", "Run without --agent/--json/--quiet or in a TTY")
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Ponto CLI")
	fmt.Fprintln(cmd.ErrOrStderr())

	globalExists := config.Exists()
	localPath := config.LocalConfigPath()
	if globalExists || localPath != "" {
		var reconfigure bool
		configLocation := "global config"
		if localPath != "" {
			configLocation = "local config (" + localPath + ")"
		}
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Existing %s found. Reconfigure?", configLocation)).
			Value(&reconfigure).
			Run()
		if err != nil {
			fmt.Println("Setup cancelled.")
			return nil //nolint:nilerr // user cancelled prompt
		}
		if !reconfigure {
			fmt.Println("Setup cancelled. Existing configuration unchanged.")
			return nil
		}
	}

	apiURL := ""
	if cfg != nil {
		apiURL = cfg.APIURL
	}
	if apiURL == "" {
		apiURL = "https://ponto.example.com"
	}
	err := huh.NewInput().
		Title("Enter your Ponto URL").
		Placeholder("https://ponto.example.com").
		Value(&apiURL).
		Validate(validateAPIURL).
		Run()
	if err != nil {
		fmt.Println("Setup cancelled.")
		return nil //nolint:nilerr // user cancelled prompt
	}

	var token string
	err = huh.NewInput().
		Title("Enter your API token").
		Description("Generate one in Ponto preferences.").
		Placeholder("ponto_...").
		Value(&token).
		EchoMode(huh.EchoModePassword).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("token is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		fmt.Println("Setup cancelled.")
		return nil //nolint:nilerr // user cancelled prompt
	}

	profileName := cfgProfile
	if profileName == "" && cfg != nil {
		profileName = cfg.Profile
	}
	if profileName == "" {
		profileName = "default"
	}

	var saveGlobal bool
	err = huh.NewSelect[bool]().
		Title("Where should we save the configuration?").
		Options(
			huh.NewOption("Global (~/.config/ponto/config.yaml)", true),
			huh.NewOption("Local (.ponto.yaml in current directory)", false),
		).
		Value(&saveGlobal).
		Run()
	if err != nil {
		fmt.Println("Setup cancelled.")
		return nil //nolint:nilerr // user cancelled prompt
	}

	newConfig := &config.Config{
		Token:   token,
		Profile: profileName,
		APIURL:  apiURL,
	}

	if saveGlobal {
		credstoreSaved := false
		if creds != nil {
			if err := credsSaveProfileToken(profileName, token); err != nil {
				fmt.Printf("Warning: could not save token to credential store: %v\n", err)
			} else {
				credstoreSaved = true
			}
		}

		ensureProfile(profileName, apiURL, "")
		if profiles != nil {
			_ = profiles.SetDefault(profileName)
		}

		existingConfig := config.LoadGlobal()
		existingConfig.Profile = profileName
		existingConfig.APIURL = apiURL
		if credstoreSaved {
			existingConfig.Token = ""
		} else {
			existingConfig.Token = token
		}
		if err := existingConfig.Save(); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("Configuration saved to ~/.config/ponto/config.yaml")
	} else {
		if err := newConfig.SaveLocal(); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("Configuration saved to .ponto.yaml")
		fmt.Println()
		fmt.Println("Remember to add .ponto.yaml to your .gitignore to avoid committing your token.")
	}

	if err := setupAgents(cmd); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're all set. Try: ponto commands")
	return nil
}

// validateAPIURL checks that the URL is well-formed and enforces HTTPS for non-localhost URLs.
func validateAPIURL(s string) error {
	if s == "" {
		return fmt.Errorf("URL is required")
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("URL must include a hostname")
	}
	if u.Scheme == "http" {
		host := u.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" && !strings.HasSuffix(host, ".localhost") {
			return fmt.Errorf("non-localhost URLs must use https://")
		}
	}
	return nil
}
