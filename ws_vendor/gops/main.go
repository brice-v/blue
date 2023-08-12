// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program gops is a tool to list currently running Go processes.
package gops

import (
	"strconv"

	"blue/ws_vendor/gops/internal/cmd"
)

func Gops_main(args []string) error {
	var root = cmd.NewRoot()
	root.AddCommand(cmd.ProcessCommand())
	root.AddCommand(cmd.TreeCommand())
	root.AddCommand(cmd.AgentCommands()...)

	// Legacy support for `gops <pid>` command.
	//
	// When the second argument is provided as int as opposed to a sub-command
	// (like proc, version, etc), gops command effectively shortcuts that
	// to `gops process <pid>`.
	if len(args) > 1 {
		// See 1st argument appears to be a pid rather than a subcommand
		_, err := strconv.Atoi(args[0])
		if err == nil {
			err = cmd.ProcessInfo(args[:])
			return err
		}
	}

	return root.Execute()
}
