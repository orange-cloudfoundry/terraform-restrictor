package main

import (
	"github.com/hashicorp/terraform/terraform"
	"path/filepath"
	"strings"
	"os/user"
	"fmt"
)

func DiffActionToMethod(action terraform.DiffChangeType) Method {
	var currentMethod Method
	switch action {
	case terraform.DiffCreate:
		currentMethod = Create
	case terraform.DiffUpdate:
		currentMethod = Update
	case terraform.DiffDestroy:
		currentMethod = Delete
	case terraform.DiffDestroyCreate:
		currentMethod = Delete
	default:
		currentMethod = None
	}
	return currentMethod
}

func expandPath(path string) (string, error) {

	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("Getting current user home dir: %s", err.Error())
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Getting absolute path: %s", err.Error())
	}

	return path, nil
}
