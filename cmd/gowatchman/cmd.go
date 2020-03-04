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

	var (
		js         = cmd.Flags().StringP("json-command", "j", "", "")
		persistent = cmd.Flags().BoolP("persistent", "p", false, "")
	)

	cmd.RunE = func(*cobra.Command, []string) error {
		if *js != "" {
			if *persistent {
				return doSendPersistent(*js)
			}
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

func doSendPersistent(js string) error {
	cl := &watchman.Client{}
	var in []interface{}
	if err := json.Unmarshal([]byte(js), &in); err != nil {
		return err
	}

	var (
		out interface{}
		ch  = make(chan interface{})
	)

	go func() {
		for msg := range ch {
			b, _ := json.MarshalIndent(msg, "", "    ")
			_, _ = fmt.Fprint(os.Stdout, string(b)+"\n")
		}
	}()

	if err := cl.Send(&out, in...); err != nil {
		return err
	}

	stop, err := cl.Receive(ch)
	defer stop()
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, string(b)+"\n")
	if err != nil {
		return err
	}

	select {}
}
