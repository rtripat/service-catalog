/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plan

import (
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/output"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	*command.Context
	lookupByUUID bool
	showSchemas  bool
	uuid         string
	name         string
}

// NewDescribeCmd builds a "svcat describe plan" command
func NewDescribeCmd(cxt *command.Context) *cobra.Command {
	describeCmd := &describeCmd{Context: cxt}
	cmd := &cobra.Command{
		Use:     "plan NAME",
		Aliases: []string{"plans", "pl"},
		Short:   "Show details of a specific plan",
		Example: command.NormalizeExamples(`
  svcat describe plan standard800
  svcat describe plan --uuid 08e4b43a-36bc-447e-a81f-8202b13e339c
`),
		PreRunE: command.PreRunE(describeCmd),
		RunE:    command.RunE(describeCmd),
	}
	cmd.Flags().BoolVarP(
		&describeCmd.lookupByUUID,
		"uuid",
		"u",
		false,
		"Whether or not to get the class by UUID (the default is by name)",
	)
	cmd.Flags().BoolVarP(
		&describeCmd.showSchemas,
		"show-schemas",
		"",
		true,
		"Whether or not to show instance and binding parameter schemas",
	)
	return cmd
}

func (c *describeCmd) Validate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("a plan name or uuid is required")
	}

	if c.lookupByUUID {
		c.uuid = args[0]
	} else {
		c.name = args[0]
	}

	return nil
}

func (c *describeCmd) Run() error {
	return c.describe()
}

func (c *describeCmd) describe() error {
	var plan *v1beta1.ClusterServicePlan
	var err error
	if c.lookupByUUID {
		plan, err = c.App.RetrievePlanByID(c.uuid)
	} else if strings.Contains(c.name, "/") {
		names := strings.Split(c.name, "/")
		if len(names) != 2 {
			return fmt.Errorf("failed to parse class/plan name combination '%s'", c.name)
		}
		plan, err = c.App.RetrievePlanByClassAndPlanNames(names[0], names[1])
	} else {
		plan, err = c.App.RetrievePlanByName(c.name)
	}
	if err != nil {
		return err
	}

	// Retrieve the class as well because plans don't have the external class name
	class, err := c.App.RetrieveClassByPlan(plan)
	if err != nil {
		return err
	}

	output.WritePlanDetails(c.Output, plan, class)

	instances, err := c.App.RetrieveInstancesByPlan(plan)
	if err != nil {
		return err
	}
	output.WriteAssociatedInstances(c.Output, instances)

	if c.showSchemas {
		output.WritePlanSchemas(c.Output, plan)
	}

	return nil
}
