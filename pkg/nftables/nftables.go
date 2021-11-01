package nftables

import (
	"fmt"

	"github.com/b-m-f/wg-pod/pkg/shell"
)

type Protocol int

const (
	TCP Protocol = iota + 1
	UDP
)

type Port struct {
	Number   uint16
	Protocol Protocol
}

type PortMap struct {
	Interface Port
	Container Port
}

func CreatePortMappings(namespace string, interfaceName string, portMapping []PortMap) error {
	// nft 'add chain nat prerouting { type nat hook prerouting priority -100; }'
	// nft add rule nat prerouting tcp dport $INTERFACE_PORT redirect to $CONTAINER_PORT

	_, err := shell.ExecuteCommand("ip", []string{"netns", "exec", namespace, "nft", "add", "table", "nat"})
	if err != nil {
		return err
	}

	_, err = shell.ExecuteCommand("ip", []string{"netns", "exec", namespace, "nft",
		"add chain nat prerouting { type nat hook prerouting priority -100; }"})
	if err != nil {
		return err
	}

	for _, mapping := range portMapping {
		protocol := ""
		if mapping.Interface.Protocol == TCP {
			protocol = "tcp"
		}
		if mapping.Interface.Protocol == UDP {
			protocol = "udp"
		}
		_, err = shell.ExecuteCommand("ip", []string{
			"netns", "exec", namespace, "nft",
			"add", "rule", "nat", "prerouting",
			"iifname", interfaceName,
			protocol, "dport", fmt.Sprint(mapping.Interface.Number),
			"redirect", "to", fmt.Sprint(mapping.Container.Number),
		})
		fmt.Printf("Forward port %d on WireGuard interface %s in namespace %s to container port %d\n", mapping.Interface.Number, interfaceName, namespace, mapping.Container.Number)
		if err != nil {
			return err
		}

	}
	return nil
}
