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
package wireguard

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Interface struct {
	Address    string
	PrivateKey string
}

type Peer struct {
	Endpoint   string
	AllowedIPs string
	PublicKey  string
	KeepAlive  int64
}

type Config struct {
	Interface Interface
	Peers     []Peer
}

// ON_CHANGE config
// Example Config
//
// [Interface]
// Address = 10.0.0.2
// PrivateKey = gC+ly0/V4jXu+K4k+nqiEGo/4On5wXu56FvSyj1tnkQ=

// [Peer]
// Endpoint = 1.1.1.1:11111
// AllowedIPs = 10.0.0.0/8
// PublicKey = gC+ly0/V4jXu+K4k+nqiEGo/4On5wXu56FvSyj1tnkQ=
// PersistentKeepalive = 25

func GetConfig(path string) (Config, error) {
	config := Config{Interface: Interface{}, Peers: []Peer{}}
	currentSection := ""
	currentPeer := -1

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		stringParts := strings.Split(strings.TrimSpace(line), " ")
		// an empty string will be returned as string[""] by the strings.Split function
		if stringParts[0] == "" {
			continue
		}

		if len(stringParts) > 2 {
			key := stringParts[0]
			if currentSection == "" {
				return Config{}, fmt.Errorf("got %v. Expected [Interface] or [Peer]", stringParts)
			}
			switch currentSection {
			case "interface":
				switch key {
				case "Address":
					config.Interface.Address = stringParts[2]
				case "PrivateKey":
					config.Interface.PrivateKey = stringParts[2]
				}

			case "peer":
				switch key {
				case "Endpoint":
					config.Peers[currentPeer].Endpoint = stringParts[2]
				case "AllowedIPs":
					config.Peers[currentPeer].AllowedIPs = stringParts[2]
				case "PublicKey":
					config.Peers[currentPeer].PublicKey = stringParts[2]
				case "PersistentKeepalive":
					keepAlive, err := strconv.ParseInt(stringParts[2], 10, 64)
					if err != nil {
						return Config{}, errors.New("invalid value for the PersistentKeepAlive")
					}
					config.Peers[currentPeer].KeepAlive = keepAlive

				}

			}
		} else {
			switch section := strings.Replace(strings.Replace(stringParts[0], "[", "", -1), "]", "", -1); section {
			case "Interface":
				currentSection = strings.ToLower(section)
			case "Peer":
				currentPeer = currentPeer + 1
				config.Peers = append(config.Peers, Peer{})
				currentSection = strings.ToLower(section)
			default:
				return Config{}, fmt.Errorf("got %v. Expected [Interface] or [Peer]", section)
			}

		}

	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return config, nil
}
