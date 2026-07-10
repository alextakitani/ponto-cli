package commands

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

func projectFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Project name")
	cmd.Flags().Int("client", 0, "Client ID")
	cmd.Flags().String("color", "", "Project color (#rrggbb)")
	cmd.Flags().Int("rate-cents", 0, "Hourly rate override in cents")
}

var projectDefaultCmd = &cobra.Command{
	Use:   "default [ID]",
	Short: "Set or clear the default project",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clear, _ := cmd.Flags().GetBool("clear")
		c, err := domainClient()
		if err != nil {
			return err
		}
		if clear {
			id := ""
			if len(args) > 0 {
				id = args[0]
			} else {
				resp, err := c.Get("/projects")
				if err != nil {
					return err
				}
				for _, item := range toSliceAny(resp.Data) {
					project, ok := item.(map[string]any)
					if !ok || project["default"] != true {
						continue
					}
					if n, ok := intFromAny(project["id"]); ok {
						id = strconv.FormatInt(n, 10)
						break
					}
				}
				if id == "" {
					printMutation(nil, "Default project cleared", nil)
					return nil
				}
			}
			if _, err := c.Delete("/projects/" + id + "/default"); err != nil {
				return err
			}
			printMutation(nil, "Default project cleared", nil)
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("project ID is required unless --clear is used")
		}
		resp, err := c.Post("/projects/"+args[0]+"/default", nil)
		if err != nil {
			return err
		}
		printMutation(resp.Data, "Default project set", nil)
		return nil
	},
}

func init() {
	projectCmd := newCatalogCmd("project", "projects", "project", projectColumns, projectFlags, "name")
	list, _, err := projectCmd.Find([]string{"list"})
	if err == nil && list != nil {
		list.Flags().Int("client", 0, "Filter by client ID")
		list.RunE = func(cmd *cobra.Command, args []string) error {
			if err := checkLimitAll(fetchAll(cmd)); err != nil {
				return err
			}
			c, err := domainClient()
			if err != nil {
				return err
			}
			values := url.Values{}
			applyCommonListParams(cmd, values)
			if changed(cmd, "client") {
				v, _ := cmd.Flags().GetInt("client")
				values.Set("client_id", strconv.Itoa(v))
			}
			resp, err := c.GetWithPagination(queryPath("/projects", values), fetchAll(cmd))
			if err != nil {
				return err
			}
			printCollection(resp, enrichPresentation(resp.Data), projectColumns, "projects", fetchAll(cmd), nil)
			return nil
		}
	}
	projectDefaultCmd.Flags().Bool("clear", false, "Clear the default project")
	projectCmd.AddCommand(projectDefaultCmd)
	rootCmd.AddCommand(projectCmd)
}
