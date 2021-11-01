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
package join

import (
	"fmt"
	"os"

	"github.com/b-m-f/wg-pod/pkg/nftables"
	"github.com/b-m-f/wg-pod/pkg/podman"
	"github.com/b-m-f/wg-pod/pkg/shell"
	"github.com/b-m-f/wg-pod/pkg/wireguard"
)

func JoinContainerIntoNetwork(name string, pathToConfig string, portMappings []nftables.PortMap) error {

	namespace, err := podman.GetNamespace(name)
	if err != nil {
		return fmt.Errorf("%s\n %s", "Problem when trying to determine the Network namespace", err.Error())
	}
	// Get the Network Namespace that Podman has set up for the container
	fmt.Printf("Adding container %s into WireGuard network defined in %s\n", name, pathToConfig)

	config, err := wireguard.GetConfig(pathToConfig)
	if err != nil {
		return fmt.Errorf("%s\n %s", "Problem when trying to read the wg-quick config", err.Error())
	}

	// Create a temporary private key file
	os.MkdirAll("/run/containers/network", 0700)
	privateKeyPath := "/run/containers/network/" + name + ".pkey"
	if err := os.WriteFile(privateKeyPath, []byte(config.Interface.PrivateKey), 0600); err != nil {
		return fmt.Errorf("%s\n %s", "problem when creating a temporary key file for the WireGuard interface", err.Error())
	}
	fmt.Printf("Create temporary private key file for WireGuard interface at %s \n", privateKeyPath)

	// Add a new Wireguard interface with the name of the container
	_, err = shell.ExecuteCommand("ip", []string{"link", "add", name, "type", "wireguard"})
	if err != nil {
		return fmt.Errorf("%s\n %s", "Problem when trying to create the new interface", err.Error())
	}
	fmt.Printf("Added new WireGuard interface %s\n", name)

	// Move the interface into the Network Namespace created by Podman
	_, err = shell.ExecuteCommand("ip", []string{"link", "set", name, "netns", namespace})
	if err != nil {
		return fmt.Errorf("problem when moving the WireGuard interface %s into the container namespace %s \n%s", name, namespace, err.Error())
	}
	fmt.Printf("Added WireGuard interface %s to network namespace %s\n", name, namespace)

	// Set the IP address of the WireGuard interface
	_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "addr", "add", config.Interface.Address, "dev", name})
	if err != nil {
		return fmt.Errorf("problem when setting the IP address %s for the WireGuard interface %s in namespace %s,\n %s", config.Interface.Address, name, namespace, err.Error())
	}
	fmt.Printf("Set IP address of WireGuard interface %s in namespace %s to %s\n", name, namespace, config.Interface.Address)

	// Set the config onto the Interface
	arguments := []string{"netns", "exec", namespace, "wg", "set", name, "private-key", privateKeyPath}
	for _, peer := range config.Peers {
		arguments = append(arguments, "peer", peer.PublicKey, "allowed-ips", peer.AllowedIPs, "endpoint", peer.Endpoint, "persistent-keepalive", fmt.Sprint(peer.KeepAlive))
	}
	_, err = shell.ExecuteCommand("ip", arguments)
	if err != nil {
		return fmt.Errorf("problem when configuring WireGuard interface %s in namespace %s with config %s\n%s", name, namespace, pathToConfig, err.Error())
	}
	fmt.Printf("Set Config %s onto WireGuard interface %s in namespace %s\n", pathToConfig, name, namespace)

	//## Set the interface active
	_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "link", "set", name, "up"})
	if err != nil {
		return fmt.Errorf("problem when activating the WireGuard interface %s in namespace %s\n%s", name, namespace, err.Error())
	}

	fmt.Printf("Activated WireGuard interface %s in namespace %s \n", name, namespace)

	//## Set a new route for all peers AllowedIPs to go over the WireGuard interface
	for _, peer := range config.Peers {
		_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "route", "add", peer.AllowedIPs, "dev", name})
		if err != nil {
			return fmt.Errorf("problem when setting the default route in %s to go through %s\n %s", namespace, name, err.Error())
		}
		fmt.Printf("Route %s in namespace %s through WireGuard interface %s \n", peer.AllowedIPs, namespace, name)
	}

	// Set up port mapping if provided
	if len(portMappings) > 0 {
		err := nftables.CreatePortMappings(namespace, name, portMappings)
		if err != nil {
			return err
		}
	}

	return nil
}
