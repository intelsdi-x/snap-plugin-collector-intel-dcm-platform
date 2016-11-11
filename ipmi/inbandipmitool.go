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
	"os"
	"sync"
)

// LinuxInBandIpmitool implements communication with ipmitool on linux
type LinuxInBandIpmitool struct {
	Device   string
	Channel  string
	Slave    string
	Protocol string
	mutex    sync.Mutex
}

// BatchExecRaw performs batch of requests to given device.
// Returns array of responses in order corresponding to requests.
// Error is returned when any of requests failed.
func (al *LinuxInBandIpmitool) BatchExecRaw(requests []IpmiRequest, host string) ([]IpmiResponse, error) {

	results := make([]IpmiResponse, len(requests))

	for i, r := range requests {
		results[i].Data = ExecIpmiToolLocal(r.Data, al, true)
		results[i].IsValid = 1
	}

	return results, nil

}

func (al *LinuxInBandIpmitool) ExecRaw(request IpmiRequest, host string) (*IpmiResponse, error) {

	results := make([]IpmiResponse, 1)

	results[0].Data = ExecIpmiToolLocal(request.Data, al, false)
	results[0].IsValid = 1

	return &results[0], nil
}

// GetPlatformCapabilities returns host NM capabilities
func (al *LinuxInBandIpmitool) GetPlatformCapabilities(requests []RequestDescription, _ []string) map[string][]RequestDescription {
	host, _ := os.Hostname()
	validRequests := make(map[string][]RequestDescription, 0)
	validRequests[host] = make([]RequestDescription, 0)

	for _, request := range requests {
		response := ExecIpmiToolLocal(request.Request.Data, al, true)
		j := 0

		for i := range response {
			if response[i] == 0 {
				j++
			}
		}
		if j != len(response) {
			validRequests[host] = append(validRequests[host], request)
		}
	}
	if al.Protocol == "dcmi" {
		// check thermal capability
		resp := ExecIpmiToolLocal(CmdDCMIThermalCap, al, false)
		if len(resp) > 5 {
			data := resp[1:]
			//sdrId := data[4] + uint16(data[5])<<8
			CmdSDR[4] = data[3]
			CmdSDR[5] = data[4]
			resp := ExecIpmiToolLocal(CmdSDR, al, false)
			if len(resp) > 10 {
				data = resp[1:]
				DcmiThermal.Request.Data[2] = data[9]
				validRequests[host] = append(validRequests[host], DcmiThermal)
			}
		}
	}
	return validRequests

}
