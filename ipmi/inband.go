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
	"sync"
	"unsafe"
)

// #include "linux_inband.h"
import "C"

// LinuxInband Implements communication with openipmi driver on linux
type LinuxInband struct {
	Device   string
	Protocol string
	mutex    sync.Mutex
}

// BatchExecRaw Performs batch of requests to given device.
// Returns array of responses in order corresponding to requests.
// Error is returned when any of requests failed.
func (al *LinuxInband) BatchExecRaw(requests []IpmiRequest, host string) ([]IpmiResponse, error) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	n := len(requests)
	info := C.struct_IpmiStatusInfo{}
	inputs := make([]C.struct_IpmiCommandInput, n)
	outputs := make([]C.struct_IpmiCommandOutput, n)

	for i, r := range requests {
		for j, b := range r.Data {
			inputs[i].data[j] = C.char(b)
		}
		inputs[i].data_len = C.int(len(r.Data))
		inputs[i].channel = C.short(r.Channel)
		inputs[i].slave = C.uchar(r.Slave)
	}

	protocol := al.Protocol
	var errcode C.int
	if protocol == "node_manager" {
		errcode = C.IPMI_BatchCommands(C.CString(al.Device), &inputs[0], &outputs[0],
			C.int(n), C.int(3), &info)
	} else {
		errcode = C.IPMI_System_BatchCommands(C.CString(al.Device), &inputs[0], &outputs[0],
			C.int(n), C.int(3), &info)
	}

	switch {
	case errcode < 0:
		return nil, fmt.Errorf("%d : Invalid call", errcode)
	case errcode > 0:
		return nil, fmt.Errorf("%d : System error [%d : %s]", errcode,
			info.system_error, C.GoString(&info.error_str[0]))
	}

	results := make([]IpmiResponse, n)

	for i, r := range outputs {
		results[i].Data = C.GoBytes(unsafe.Pointer(&r.data[0]), r.data_len)
		results[i].IsValid = uint(r.is_valid)
	}

	return results, nil
}

// GetPlatformCapabilities returns host NM capabilities
func (al *LinuxInband) GetPlatformCapabilities(requests []RequestDescription, hostlist []string) map[string][]RequestDescription {

	host := hostlist[0]

	validRequests := make(map[string][]RequestDescription, 0)
	validRequests[host] = make([]RequestDescription, 0)

	requestCmds := make([]IpmiRequest, len(requests))
	for i, _ := range requests {
		requestCmds[i] = requests[i].Request
	}

	response, e := al.BatchExecRaw(requestCmds, host)
	if e == nil {

		for i, r := range response {
			if r.IsValid == 1 {
				validRequests[host] = append(validRequests[host], requests[i])
			}
		}

	}

	if al.Protocol == "dcmi" {
		// check thermal capability
		dcmiRequest := IpmiRequest{Data: CmdDCMIThermalCap, Channel: 0x0, Slave: 0x0}
		resp, e := al.ExecRaw(dcmiRequest, host)
		if e == nil && len(resp.Data) > 5 {
			data := resp.Data[1:]
			//sdrId := data[4] + uint16(data[5])<<8
			dcmiRequest.Data = make([]byte, len(CmdSDR))
			copy(dcmiRequest.Data, CmdSDR)
			dcmiRequest.Data[4] = data[3]
			dcmiRequest.Data[5] = data[4]
			resp, e = al.ExecRaw(dcmiRequest, host)

			if e == nil && len(resp.Data) > 10 {
				data = resp.Data[1:]
				DcmiThermal.Request.Data[2] = data[9]
				validRequests[host] = append(validRequests[host], DcmiThermal)
			}
		}
	}
	return validRequests
}

func (al *LinuxInband) ExecRaw(request IpmiRequest, host string) (*IpmiResponse, error) {
	n := 1
	info := C.struct_IpmiStatusInfo{}
	inputs := make([]C.struct_IpmiCommandInput, 1)
	outputs := make([]C.struct_IpmiCommandOutput, 1)

	for j, b := range request.Data {
		inputs[0].data[j] = C.char(b)
	}

	inputs[0].data_len = C.int(len(request.Data))
	inputs[0].channel = C.short(request.Channel)
	inputs[0].slave = C.uchar(request.Slave)

	errcode := C.IPMI_System_BatchCommands(C.CString(al.Device), &inputs[0], &outputs[0],
		C.int(n), C.int(3), &info)

	switch {
	case errcode < 0:
		return nil, fmt.Errorf("%d : Invalid call", errcode)
	case errcode > 0:
		return nil, fmt.Errorf("%d : System error [%d : %s]", errcode,
			info.system_error, C.GoString(&info.error_str[0]))
	}

	results := make([]IpmiResponse, n)

	for i, r := range outputs {
		results[i].Data = C.GoBytes(unsafe.Pointer(&r.data[0]), r.data_len)
		results[i].IsValid = uint(r.is_valid)
	}

	return &results[0], nil
}
