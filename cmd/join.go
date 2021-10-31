/*
Copyright © 2021 b-m-f<max@ehlers.berlin>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"errors"

	"github.com/b-m-f/wg-pod/pkg/join"
	"github.com/spf13/cobra"
)

var ContainerName string
var ConfigPath string

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join [name] [path]",
	Short: "Add a container/pod into a WireGuard network",
	Long: `Add a container/pod into a WireGuard network

[name]: Container Name
[path]: absolute path to the WireGuard config`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires at least two arg")
		}
		// TODO: check that input 1 matches an existing podman container
		// podman.containerExists
		// TODO: check that input 2 matches a valid config. Check that the Path exists and it has all valid fields
		// wireguard.validateConfig
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := join.JoinContainerIntoNetwork(args[0], args[1])
		if err != nil {
			return err

		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)
}
