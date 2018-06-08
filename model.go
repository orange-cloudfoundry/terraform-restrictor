package main

import (
	"regexp"
	"fmt"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/go-multierror"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

const (
	Create Method = "create"
	Update Method = "update"
	Delete Method = "delete"
	None   Method = "none"
)

type Method string

func (m *Method) Method(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	method := Method(s)
	if method != Create && method != Update && method != Delete {
		return fmt.Errorf("invalid method given '%s'", method)
	}
	*m = method
	return nil
}

type Methods []Method

func (ms Methods) Match(currentMethod Method) bool {
	for _, m := range ms {
		if m == currentMethod {
			return true
		}
	}
	return false
}

type Regexp struct {
	*regexp.Regexp
	Raw string
}

func (re *Regexp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	regex, err := regexp.Compile("^(?:" + s + ")$")
	if err != nil {
		return err
	}
	re.Raw = s
	re.Regexp = regex
	return nil
}

type Regexps []Regexp

func (r Regexps) MatchString(s string) bool {
	if len(r) == 0 {
		return true
	}
	for _, match := range r {
		if match.MatchString(s) {
			return true
		}
	}
	return false
}

func (r Regexps) String() string {
	if len(r) == 1 {
		return fmt.Sprintf("/^%s$/", r[0].Raw)
	}
	inline := make([]string, len(r))
	for i, match := range r {
		inline[i] = fmt.Sprintf("(%s)", match.Raw)
	}
	return fmt.Sprintf("/^%s$/", strings.Join(inline, "|"))
}

type CheckAttrs []CheckAttr

func (cs CheckAttrs) Check(fPlan *format.InstanceDiff) error {
	var result error
	for _, c := range cs {
		err := c.Check(fPlan)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

type CheckAttr struct {
	Path         Regexp  `yaml:"key"`
	Validate     Regexps `yaml:"validate"`
	Unauthorized Methods `yaml:"unauthorized"`
}

func (c CheckAttr) Check(fPlan *format.InstanceDiff) error {
	var err error
	attrs := fPlan.Attributes
	var finalAttr *format.AttributeDiff
	for _, attr := range attrs {
		if !c.Path.MatchString(attr.Path) {
			continue
		}
		finalAttr = attr
	}
	if finalAttr == nil {
		return nil
	}
	m := DiffActionToMethod(finalAttr.Action)
	if c.Unauthorized.Match(m) {
		err = multierror.Append(err, fmt.Errorf("Attribute '%s' is unauthorized to use method '%s'", finalAttr.Path, m))
	}
	if m == Delete || m == None {
		return err
	}
	if !c.Validate.MatchString(finalAttr.NewValue) {
		err = multierror.Append(
			err,
			fmt.Errorf("Attribute '%s' has invalid value, it should match regex '%s' and it was '%s'",
				finalAttr.Path,
				c.Validate,
				finalAttr.NewValue,
			),
		)
	}
	return err
}

type Resources []Resource

func (rs Resources) Check(fPlan *format.InstanceDiff) error {
	var result error
	for _, r := range rs {
		err := r.Check(fPlan)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

// This function use panic instead of giving an error because a flag with a default tag do not give back error on go-flags
// this is tedious to change this behaviour on go-flags
func (a *Resources) UnmarshalFlag(data string) error {

	var b []byte
	var err error

	absPath, err := expandPath(data)
	if err != nil {
		panic(fmt.Errorf("Getting absolute path '%s': %s", data, err.Error()))
	}
	b, err = ioutil.ReadFile(absPath)
	if err != nil {
		panic(err)
	}

	var resources []Resource
	err = yaml.Unmarshal(b, &resources)
	if err != nil {
		panic(err)
	}
	*a = resources

	return nil
}

type Resource struct {
	Type         Regexp     `yaml:"type"`
	Name         Regexp     `yaml:"name,omitempty"`
	CheckAttrs   CheckAttrs `yaml:"check_attrs"`
	Unauthorized Methods    `yaml:"unauthorized"`
}

func (r Resource) Check(fPlan *format.InstanceDiff) error {
	if !r.Type.MatchString(fPlan.Addr.Type) {
		return nil
	}
	if r.Name.Raw != "" && !r.Name.MatchString(fPlan.Addr.Name) {
		return nil
	}
	var result error
	m := DiffActionToMethod(fPlan.Action)
	if r.Unauthorized.Match(m) {
		result = multierror.Append(result, fmt.Errorf("Resource is unauthorized to use method '%s'", m))
	}

	err := r.CheckAttrs.Check(fPlan)
	if err != nil {
		result = multierror.Append(result, err)
	}
	return result
}
