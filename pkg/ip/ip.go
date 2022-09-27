
package ip

import (
	"fmt"
	"net"

	"github.com/b-m-f/wg-pod/pkg/shell"
)

type Route struct {
	Target net.IP
	Gateway net.IP
}


func CreateRoute(namespace string, additionalRoutes []Route) error {

	for _, route := range additionalRoutes{
		_, err := shell.ExecuteCommand("ip", []string{
			"netns", "exec", namespace, "ip",
			"route", "add", fmt.Sprint(route.Target.To4()), "via", fmt.Sprint(route.Gateway.To4()),
		})
		fmt.Printf("Routing to %s via %s in namespace %s\n", route.Target.To4(), route.Gateway.To4(), namespace)
		if err != nil {
			return err
		}

	}
	return nil
}