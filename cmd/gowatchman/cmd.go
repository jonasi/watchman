package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jonasi/watchman"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use: "gowatchman",
	}

	js := cmd.Flags().StringP("json-command", "j", "", "")

	cmd.RunE = func(*cobra.Command, []string) error {
		if *js != "" {
			return doSend(*js)
		}

		return cmd.Usage()
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func doSend(js string) error {
	cl := &watchman.Client{}
	var in []interface{}
	if err := json.Unmarshal([]byte(js), &in); err != nil {
		return err
	}

	var out interface{}
	if err := cl.Send(&out, in...); err != nil {
		return err
	}

	b, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, string(b)+"\n")
	return err
}
