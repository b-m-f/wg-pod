# WireGuard config

This is an example configuration to use with `wg-pod`.

It is very similar to the [wg-quick](https://git.zx2c4.com/wireguard-tools/about/src/man/wg-quick.8) config.

This means you can use any existing `wg-quick` config with `wg-pod`.
Only necessary keys are extracted and all other are disregarded.

<!-- ON_CHANGE config -->

```
[Interface]
Address = 10.0.0.2
PrivateKey = gC+ly0/V4jXu+K4k+nqiEGo/4On5wXu56FvSyj1tnkQ=

[Peer]
Endpoint = 1.1.1.1:11111
AllowedIPs = 10.0.0.0/8
PublicKey = gC+ly0/V4jXu+K4k+nqiEGo/4On5wXu56FvSyj1tnkQ=
PersistentKeepalive = 25
```

