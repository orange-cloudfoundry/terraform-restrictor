package main

import (
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/command/format"
	"os"
	"io"
	"github.com/hashicorp/go-multierror"
	"strings"
)

type Plan struct {
	*format.Plan
}

func (a *Plan) UnmarshalFlag(data string) error {
	if data == "" {
		return nil
	}

	var f io.Reader
	var err error

	if data == "-" {
		f = os.Stdin
		if err != nil {
			return fmt.Errorf("Reading from stdin: %s", err.Error())
		}
	} else {
		absPath, err := expandPath(data)
		if err != nil {
			return fmt.Errorf("Getting absolute path '%s': %s", data, err.Error())
		}
		f, err = os.Open(absPath)
		if err != nil {
			return fmt.Errorf("Opening file '%s': %s", absPath, err.Error())
		}
	}
	plan, err := terraform.ReadPlan(f)
	if err != nil {
		return fmt.Errorf("Reading plan: %s", err.Error())
	}
	a.Plan = format.NewPlan(plan)
	return nil
}

type PlanArg struct {
	Plan Plan `positional-arg-name:"PATH" description:"Path to a terraform plan file (use - to load plan from stdin)"`
}

type RestrictorFlag struct {
	PlanArg      PlanArg   `positional-args:"true" required:"1"`
	Restrictions Resources `long:"file" short:"f" required:"true" default:"restrictions.yml" description:"Path to the restrictions definition yaml file"`
	Verbose      bool      `short:"v" long:"verbose" description:"Verbose output"`
}

type MultipleErrors []*multierror.Error

func (es MultipleErrors) Error() string {
	var result string
	for _, e := range es {
		result += fmt.Sprintf("%s\n\n", e.Error())
	}
	return "\n" + result
}

func CheckRestrictions(rFlag RestrictorFlag) error {
	errors := make(MultipleErrors, 0)
	for _, r := range rFlag.PlanArg.Plan.Resources {
		var err error
		err = rFlag.Restrictions.Check(r)
		if err == nil {
			continue
		}
		err.(*multierror.Error).ErrorFormat = CreateErrorFormat(*r)
		errors = append(errors, err.(*multierror.Error))
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

func CreateErrorFormat(i format.InstanceDiff) func(es []error) string {
	return func(es []error) string {
		points := make([]string, len(es))
		for i, err := range es {
			points[i] = fmt.Sprintf("  * %s", err)
		}

		return fmt.Sprintf(
			"- Resource '%s', %d errors:\n%s",
			i.Addr, len(es), strings.Join(points, "\n"))
	}
}
