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
	"github.com/b-m-f/wg-pod/pkg/ip"
	"github.com/b-m-f/wg-pod/pkg/podman"
	"github.com/b-m-f/wg-pod/pkg/shell"
	"github.com/b-m-f/wg-pod/pkg/uuid"
	"github.com/b-m-f/wg-pod/pkg/wireguard"
)

func JoinContainerIntoNetwork(containerName string, pathToConfig string, portMappings []nftables.PortMap, deleteDefault bool, additionalRoutes []ip.Route) error {
	uuid, err := uuid.GetUUID()
	if err != nil {
		return fmt.Errorf("problem when creating a UUID for the interface\n %s", err.Error())
	}
	// maximum length for an interface name by default is 15 bytes.
	// podman- takes 7, so get 8 additional ones from the uuid
	interfaceName := fmt.Sprintf("podman-%s", uuid[0:7])

	namespace, err := podman.GetNamespace(containerName)
	if err != nil {
		return fmt.Errorf("problem when trying to determine the Network namespace\n %s", err.Error())
	}
	// Get the Network Namespace that Podman has set up for the container
	fmt.Printf("Adding container %s into WireGuard network defined in %s\n", containerName, pathToConfig)

	config, err := wireguard.GetConfig(pathToConfig)
	if err != nil {
		return fmt.Errorf("problem when trying to read the wg-quick config\n %s", err.Error())
	}

	// Create a temporary private key file
	os.MkdirAll("/run/containers/network", 0700)
	privateKeyPath := "/run/containers/network/" + containerName + ".pkey"
	if err := os.WriteFile(privateKeyPath, []byte(config.Interface.PrivateKey), 0600); err != nil {
		return fmt.Errorf("problem when creating a temporary key file for the WireGuard interface\n %s", err.Error())
	}
	fmt.Printf("Create temporary private key file for WireGuard interface at %s \n", privateKeyPath)

	// Add a new Wireguard interface inside the container namespace
	_, err = shell.ExecuteCommand("ip", []string{"link", "add", interfaceName, "type", "wireguard"})
	if err != nil {
		return fmt.Errorf("problem when trying to create the new interface\n %s", err.Error())
	}
	fmt.Printf("Added new WireGuard interface %s\n", interfaceName)

	// Move interface into container namespace
	_, err = shell.ExecuteCommand("ip", []string{"link", "set", interfaceName, "netns", namespace})
	if err != nil {
		return fmt.Errorf("problem when trying to move WireGuard interface %s to namespace %s \n %s", interfaceName, namespace, err.Error())
	}
	fmt.Printf("Moved WireGuard interface %s to namespace %s\n", interfaceName, namespace)

	// Set the IP address of the WireGuard interface
	_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "addr", "add", config.Interface.Address, "dev", interfaceName})
	if err != nil {
		return fmt.Errorf("problem when setting the IP address %s for the WireGuard interface %s in namespace %s,\n %s", config.Interface.Address, interfaceName, namespace, err.Error())
	}
	fmt.Printf("Set IP address of WireGuard interface %s in namespace %s to %s\n", interfaceName, namespace, config.Interface.Address)

	// Set the config onto the Interface
	arguments := []string{"netns", "exec", namespace, "wg", "set", interfaceName, "private-key", privateKeyPath}
	for _, peer := range config.Peers {
		arguments = append(arguments, "peer", peer.PublicKey)
		if peer.PresharedKey != "" {
	            presharedKeyPath := "/run/containers/network/" + containerName + ".pskey"
	            if err := os.WriteFile(presharedKeyPath, []byte(peer.PresharedKey), 0600); err != nil {
	                    return fmt.Errorf("problem when creating a temporary key file for the PresharedKey\n %s", err.Error())
	            }
	    	    arguments = append(arguments, "preshared-key", presharedKeyPath)
		}
		arguments = append(arguments, "allowed-ips")

		ips := ""
		for _, ip := range peer.AllowedIPs {
			if len(ips) > 0  {
			ips = ips + "," + ip
			} else {
			ips = ips + ip
			}
		}
	        arguments = append(arguments,ips)
		if peer.Endpoint != "" {
			arguments = append(arguments, "endpoint", peer.Endpoint)
		}
		if peer.KeepAlive != 0 {
			arguments = append(arguments, "persistent-keepalive", fmt.Sprint(peer.KeepAlive))
		}
	}
	_, err = shell.ExecuteCommand("ip", arguments)
	if err != nil {
		return fmt.Errorf("problem when configuring WireGuard interface %s in namespace %s with config %s\n%s", interfaceName, namespace, pathToConfig, err.Error())
	}
	fmt.Printf("Set Config %s onto WireGuard interface %s in namespace %s\n", pathToConfig, interfaceName, namespace)

	//## Set the interface active
	_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "link", "set", interfaceName, "up"})
	if err != nil {
		return fmt.Errorf("problem when activating the WireGuard interface %s in namespace %s\n%s", interfaceName, namespace, err.Error())
	}

	fmt.Printf("Activated WireGuard interface %s in namespace %s \n", interfaceName, namespace)

	// Delete default route if desirec
	if deleteDefault {
		_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "route", "del", "default"})
		if err != nil {
			return fmt.Errorf("problem when deleting the default route in namespace %s\n%s", namespace, err.Error())
		}

		fmt.Printf("Successfully deleted the default route in namespace %s \n", namespace)

	}

	//## Set a new route for all peers AllowedIPs to go over the WireGuard interface
	for _, peer := range config.Peers {
		for _, ip:= range peer.AllowedIPs{
		_, err = shell.ExecuteCommand("ip", []string{"-n", namespace, "route", "add", ip , "dev", interfaceName})
		if err != nil {
			return fmt.Errorf("problem when setting the default route in %s to go through %s\n %s", namespace, interfaceName, err.Error())
		}
		fmt.Printf("Route %s in namespace %s through WireGuard interface %s \n", ip, namespace, interfaceName)
	  }
	}

	// Set up port mapping if provided
	if len(portMappings) > 0 {
		err := nftables.CreatePortMappings(namespace, interfaceName, portMappings)
		if err != nil {
			return err
		}
	}

	// establish additional routes	
	if len(additionalRoutes) > 0 {
		err := ip.CreateRoute(namespace, additionalRoutes)
		if err != nil {
			return err
		}
	}

	return nil
}
