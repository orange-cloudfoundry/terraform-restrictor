package main

import (
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"io"
)

var LogWriter io.Writer = os.Stderr

func Run(args []string) error {
	var rFlag RestrictorFlag
	parser := flags.NewParser(&rFlag, flags.HelpFlag|flags.PassDoubleDash)
	log.SetOutput(LogWriter)
	_, err := parser.ParseArgs(args[1:])
	if err != nil {
		return err
	}
	if rFlag.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	return CheckRestrictions(rFlag)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "terraform-restrictor: %s\n", r)
			os.Exit(1)
		}
	}()
	err := Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "terraform-restrictor: %s\n", err.Error())
		os.Exit(1)
	}
}
