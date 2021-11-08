# wg-pod

A tool to quickly join your [podman](https://podman.io/) container/pod into a WireGuard network.

## Explanation

wg-pod wires up the tools **ip**,**route**,[wg](https://git.zx2c4.com/wireguard) and podman.
It creates a WireGuard interface inside of the host namespace, moves it into the container namespace and then routes all traffic defined as `AllowedIPs` through the WireGuard interface.
The initial creation in the host namespace is done to assure that the interface can reach its Endpoint even if the default route in the container namespace is being deleted.

Existing interfaces in the namespace are not deleted and a route that is more specific than the default route in the namespace will still match.
This means that the container will be able to talk over both the WireGuard network and the original network that was created for it by podman.

# Commands

## join

### Parameters

- `container_name (required)`: specify the name of the container that should get connected into the network
- `config_path (required)`: absolute path to the [wireguard config](./docs/wireguard-config)

### Flags

- `port-remapping (optional)`: comma separated list of ports to remap from the interface to the container
- `delete-default (optional, default false)`: Remove the default route in the container namespace

# Dependencies

- Linux
- write permissions to `/run/containers`
- permissions to change the network `CAP_NET_ADMIN`
- nftables
- ip
- wireguard

## Cool automation

### systemd

Check out [quadlet](https://github.com/containers/quadlet) first to see how to easily generated systemd unit files.
Use `wg-pod` inside the `ExecStartPost` lifecycle of the quadlet container file to spawn containers into a network directly after creation.
Of course this also works with plain systemd unit files if you prefer not to use quadlet.

### All traffic through the container

If you set `AllowedIPs` to `0.0.0.0/0` your container will route all its traffic through the tunnel, but you must make sure to use the `-d` flag to delete the default route set up by podman.
Just be aware that it will still be able to talk to podman networks since these have more specific routes than the default.

Your container can now talk to other containers if it is inside a pod, but route all other traffic through a tunnel.

# Security considerations

- Make sure that no user (not even root) can edit around network configurations inside your container. (`CAP_NET_ADMIN` must not be given)
- The Host network that was set up during container creation is still reachable with routing rules more specific than the default route to the WireGuard VPN
