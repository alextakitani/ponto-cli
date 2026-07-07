package commands

import (
	"fmt"
	"net/url"

	"github.com/alextakitani/ponto-cli/internal/render"
	"github.com/spf13/cobra"
)

func newCatalogCmd(resource, plural, bodyKey string, cols render.Columns, flags func(*cobra.Command), requiredCreate ...string) *cobra.Command {
	parent := &cobra.Command{
		Use:   resource,
		Short: "Manage " + plural,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List " + plural,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := domainClient()
			if err != nil {
				return err
			}
			values := url.Values{}
			applyCommonListParams(cmd, values)
			resp, err := c.Get(queryPath("/"+plural, values))
			if err != nil {
				return err
			}
			printList(enrichPresentation(resp.Data), cols, fmt.Sprintf("%d %s", dataCount(resp.Data), plural), nil)
			return nil
		},
	}
	addCommonListFlags(listCmd)

	showCmd := &cobra.Command{
		Use:   "show ID",
		Short: "Show a " + resource,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := domainClient()
			if err != nil {
				return err
			}
			resp, err := c.Get("/" + plural + "/" + args[0])
			if err != nil {
				return err
			}
			printDetail(enrichPresentation(resp.Data), titleWord(resource)+" "+args[0], []Breadcrumb{
				breadcrumb("update", "ponto "+resource+" update "+args[0], "Update this "+resource),
				breadcrumb("delete", "ponto "+resource+" delete "+args[0], "Delete this "+resource),
			})
			return nil
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a " + resource,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireFlags(cmd, requiredCreate...); err != nil {
				return err
			}
			body := map[string]any{}
			flagsToBody(cmd, body)
			c, err := domainClient()
			if err != nil {
				return err
			}
			resp, err := c.Post("/"+plural, map[string]any{bodyKey: body})
			if err != nil {
				return err
			}
			printMutation(resp.Data, titleWord(resource)+" created", []Breadcrumb{
				breadcrumb("list", "ponto "+resource+" list", "List "+plural),
			})
			return nil
		},
	}
	flags(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update a " + resource,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]any{}
			flagsToBody(cmd, body)
			c, err := domainClient()
			if err != nil {
				return err
			}
			resp, err := c.Patch("/"+plural+"/"+args[0], map[string]any{bodyKey: body})
			if err != nil {
				return err
			}
			printMutation(resp.Data, titleWord(resource)+" updated", []Breadcrumb{
				breadcrumb("show", "ponto "+resource+" show "+args[0], "Show this "+resource),
			})
			return nil
		},
	}
	flags(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete ID",
		Short: "Delete a " + resource,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := domainClient()
			if err != nil {
				return err
			}
			if _, err := c.Delete("/" + plural + "/" + args[0]); err != nil {
				return err
			}
			printMutation(nil, titleWord(resource)+" deleted", []Breadcrumb{
				breadcrumb("list", "ponto "+resource+" list", "List "+plural),
			})
			return nil
		},
	}

	parent.AddCommand(listCmd, showCmd, createCmd, updateCmd, deleteCmd)
	return parent
}

func flagsToBody(cmd *cobra.Command, body map[string]any) {
	addStringIfChanged(cmd, body, "name", "name")
	addStringIfChanged(cmd, body, "currency", "currency")
	addStringIfChanged(cmd, body, "note", "note")
	addStringIfChanged(cmd, body, "color", "color")
	addIntIfChanged(cmd, body, "rate-cents", "rate_cents")
	if changed(cmd, "client") {
		v, _ := cmd.Flags().GetInt("client")
		body["client_id"] = v
	}
}
