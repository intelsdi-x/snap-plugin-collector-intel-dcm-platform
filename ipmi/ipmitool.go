// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// ExecIpmiToolLocal method runs ipmitool command on a local system
func ExecIpmiToolLocal(request []byte, strct *LinuxInBandIpmitool, isBridged bool) []byte {
	c, err := exec.LookPath("ipmitool")
	if err != nil {
		log.Debug("Unable to find ipmitool")
		return nil
	}

	var stringRequest []string
	if isBridged && strct.Slave != "0" {
		stringRequest = []string{"-b", strct.Channel, "-t", strct.Slave, "raw"}
	} else {
		stringRequest = []string{"raw"}
	}

	for i := range request {
		stringRequest = append(stringRequest, fmt.Sprintf("0x%02x", request[i]))
	}

	ret, err := exec.Command(c, stringRequest...).CombinedOutput()
	if err != nil {
		log.Debug("Unable to run ipmitool")
		return nil
	}
	returnStrings := strings.Split(string(ret), " ")
	rets := make([]byte, len(returnStrings))
	for i, element := range returnStrings {
		value, _ := strconv.ParseInt(strings.TrimSpace(element), 16, 0)
		rets[i] = byte(value)
	}

	return rets
}

// ExecIpmiToolRemote method runs ipmitool command on a remote system
func ExecIpmiToolRemote(request []byte, strct *LinuxOutOfBand, addr string, isBridged bool) []byte {
	c, err := exec.LookPath("ipmitool")
	if err != nil {
		log.Debug("Unable to find ipmitool")
		return nil
	}

	var a []string
	if isBridged && strct.Slave != "0" {
		a = []string{"-I", "lanplus", "-H", addr, "-U", strct.User, "-P", strct.Pass, "-b", strct.Channel, "-t", strct.Slave, "raw"}
	} else {
		a = []string{"-I", "lanplus", "-H", addr, "-U", strct.User, "-P", strct.Pass, "raw"}
	}
	for i := range request {
		a = append(a, fmt.Sprintf("0x%02x", request[i]))
	}

	ret, err := exec.Command(c, a...).CombinedOutput()
	if err != nil {
		log.Debug("Unable to run ipmitool")
		return nil
	}

	returnStrings := strings.Split(string(ret), " ")

	rets := make([]byte, len(returnStrings))
	for ind, el := range returnStrings {

		value, _ := strconv.ParseInt(strings.TrimSpace(el), 16, 0)
		rets[ind] = byte(value)
	}

	return rets

}
