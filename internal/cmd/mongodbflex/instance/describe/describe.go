package describe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"stackit/internal/pkg/args"
	"stackit/internal/pkg/errors"
	"stackit/internal/pkg/examples"
	"stackit/internal/pkg/globalflags"
	"stackit/internal/pkg/services/mongodbflex/client"
	"stackit/internal/pkg/tables"
	"stackit/internal/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/mongodbflex"
)

const (
	instanceIdArg = "INSTANCE_ID"
)

type inputModel struct {
	*globalflags.GlobalFlagModel
	InstanceId string
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("describe %s", instanceIdArg),
		Short: "Get details of a MongoDB Flex instance",
		Long:  "Get details of a MongoDB Flex instance",
		Args:  args.SingleArg(instanceIdArg, utils.ValidateUUID),
		Example: examples.Build(
			examples.NewExample(
				`Get details of a MongoDB Flex instance with ID "xxx"`,
				"$ stackit mongodbflex instance describe xxx"),
			examples.NewExample(
				`Get details of a MongoDB Flex instance with ID "xxx" in a table format`,
				"$ stackit mongodbflex instance describe xxx --output-format pretty"),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			model, err := parseInput(cmd, args)
			if err != nil {
				return err
			}
			// Configure API client
			apiClient, err := client.ConfigureClient(cmd)
			if err != nil {
				return err
			}

			// Call API
			req := buildRequest(ctx, model, apiClient)
			resp, err := req.Execute()
			if err != nil {
				return fmt.Errorf("read MongoDB Flex instance: %w", err)
			}

			return outputResult(cmd, model.OutputFormat, resp.Item)
		},
	}
	return cmd
}

func parseInput(cmd *cobra.Command, inputArgs []string) (*inputModel, error) {
	instanceId := inputArgs[0]

	globalFlags := globalflags.Parse(cmd)
	if globalFlags.ProjectId == "" {
		return nil, &errors.ProjectIdError{}
	}

	return &inputModel{
		GlobalFlagModel: globalFlags,
		InstanceId:      instanceId,
	}, nil
}

func buildRequest(ctx context.Context, model *inputModel, apiClient *mongodbflex.APIClient) mongodbflex.ApiGetInstanceRequest {
	req := apiClient.GetInstance(ctx, model.ProjectId, model.InstanceId)
	return req
}

func outputResult(cmd *cobra.Command, outputFormat string, instance *mongodbflex.Instance) error {
	switch outputFormat {
	case globalflags.PrettyOutputFormat:
		acls := *instance.Acl.Items
		strings.Join(acls, ",")

		table := tables.NewTable()
		table.AddRow("ID", *instance.Id)
		table.AddSeparator()
		table.AddRow("NAME", *instance.Name)
		table.AddSeparator()
		table.AddRow("STATUS", *instance.Status)
		table.AddSeparator()
		table.AddRow("STORAGE SIZE", *instance.Storage.Size)
		table.AddSeparator()
		table.AddRow("VERSION", *instance.Version)
		table.AddSeparator()
		table.AddRow("ACL", acls)
		table.AddSeparator()
		table.AddRow("FLAVOR DESCRIPTION", *instance.Flavor.Description)
		table.AddSeparator()
		table.AddRow("CPU", *instance.Flavor.Cpu)
		table.AddSeparator()
		table.AddRow("RAM", *instance.Flavor.Memory)
		table.AddSeparator()
		err := table.Display(cmd)
		if err != nil {
			return fmt.Errorf("render table: %w", err)
		}

		return nil
	default:
		details, err := json.MarshalIndent(instance, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal MongoDB Flex instance: %w", err)
		}
		cmd.Println(string(details))

		return nil
	}
}