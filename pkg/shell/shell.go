/*
copyright © 2021 b-m-f<max@ehlers.berlin>
all rights reserved.

redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

this software is provided by the copyright holders and contributors "as is"
and any express or implied warranties, including, but not limited to, the
implied warranties of merchantability and fitness for a particular purpose
are disclaimed. in no event shall the copyright holder or contributors be
liable for any direct, indirect, incidental, special, exemplary, or
consequential damages (including, but not limited to, procurement of
substitute goods or services; loss of use, data, or profits; or business
interruption) however caused and on any theory of liability, whether in
contract, strict liability, or tort (including negligence or otherwise)
arising in any way out of the use of this software, even if advised of the
possibility of such damage.
*/

package shell

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecuteCommand(app string, args []string) (string, error) {
	cmd := exec.Command(app, args...)

	var std_out bytes.Buffer
	var std_err bytes.Buffer
	cmd.Stdout = &std_out
	cmd.Stderr = &std_err
	err := cmd.Run()

	if err != nil {
		return "", fmt.Errorf("%s", std_err.String())
	}
	return std_out.String(), nil

}
