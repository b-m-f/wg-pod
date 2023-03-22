package ip

import (
	"fmt"
	"net"

	"github.com/b-m-f/wg-pod/pkg/shell"
)

type Route struct {
	Target  net.IPNet
	Gateway net.IPNet
}

func CreateRoute(namespace string, additionalRoutes []Route) error {

	for _, route := range additionalRoutes {
		_, err := shell.ExecuteCommand("ip", []string{
			"netns", "exec", namespace, "ip",
			"route", "add", fmt.Sprint(route.Target.String()), "via", fmt.Sprint(route.Gateway.String()),
		})
		fmt.Printf("Routing to %s via %s in namespace %s\n", route.Target.String(), route.Gateway.String(), namespace)
		if err != nil {
			return err
		}

	}
	return nil
}
