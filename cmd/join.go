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
	"fmt"
	"strconv"
	"strings"

	"github.com/b-m-f/wg-pod/pkg/join"
	"github.com/b-m-f/wg-pod/pkg/nftables"
	"github.com/spf13/cobra"
)

var PortMapInput string

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join [name] [path]",
	Short: "Add a container/pod into a WireGuard network",
	Long: `Add a container/pod into a WireGuard network

[name]: Container Name
[path]: absolute path to the WireGuard config

Example

wg-pod join webapp /etc/wireguard/webapp.conf --port 3030:443,3031:8080
`,
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
		portMappings := []nftables.PortMap{}
		// Split always returns array of at least len == 1
		for _, pair := range strings.Split(PortMapInput, ",") {
			portMap := nftables.PortMap{}
			if pair == "" {
				// no ports were provided
				break
			}
			ports := strings.Split(pair, ":")
			if ports[0] == "" || len(ports) < 2 {
				return fmt.Errorf("incorrect portmapping provided: %v", ports)
			}

			interfacePort, err := getPortFromString(ports[0])
			if err != nil {
				return err
			}

			containerPort, err := getPortFromString(ports[1])
			if err != nil {
				return err
			}
			portMap.Container = containerPort
			portMap.Interface = interfacePort
			portMappings = append(portMappings, portMap)
		}
		err := join.JoinContainerIntoNetwork(args[0], args[1], portMappings)
		if err != nil {
			return err

		}
		return nil
	},
}

func init() {
	joinCmd.Flags().StringVarP(&PortMapInput, "port-remapping", "p", "", "Comma separated list of PortMapping from interface into container")
	rootCmd.AddCommand(joinCmd)
}

// 80/tcp to {[port: 80, protocol: tcp]}
func getPortFromString(input string) (nftables.Port, error) {
	numberAndProtocolSplit := strings.Split(input, "/")

	port := nftables.Port{}

	if len(numberAndProtocolSplit) > 2 {
		return port, errors.New("incorrect portmapping provided")
	}
	if len(numberAndProtocolSplit) < 2 {
		castPortNumber, err := strconv.ParseUint(numberAndProtocolSplit[0], 10, 16)
		if err != nil {
			return port, fmt.Errorf("incorrect Port provided: %s. Example: 80/tcp", numberAndProtocolSplit[0])
		}
		port.Number = uint16(castPortNumber)
		port.Protocol = nftables.TCP
	} else {
		castPortNumber, err := strconv.ParseUint(numberAndProtocolSplit[0], 10, 16)
		if err != nil {
			return port, fmt.Errorf("incorrect Port provided: %s. Example: 80/tcp", numberAndProtocolSplit[0])
		}
		port.Number = uint16(castPortNumber)
		switch numberAndProtocolSplit[1] {
		case "tcp":
			port.Protocol = nftables.TCP
		case "udp":
			port.Protocol = nftables.UDP
		default:
			return port, errors.New("wrong Interface protocol provided. Valid: tcp & udp")
		}
	}
	return port, nil

}
