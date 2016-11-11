// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	log "github.com/Sirupsen/logrus"
	"sync"
	"time"
)

// LinuxOutOfBand implements communication with openipmi driver on linux
type LinuxOutOfBand struct {
	Device   string
	Channel  string
	Slave    string
	Addr     []string
	User     string
	Pass     string
	Protocol string
	mutex    sync.Mutex
}

// BatchExecRaw Performs batch of requests to given device.
// Returns array of responses in order corresponding to requests.
// Error is returned when any of requests failed.
func (al *LinuxOutOfBand) BatchExecRaw(requests []IpmiRequest, host string) ([]IpmiResponse, error) {
	var wg sync.WaitGroup
	wg.Add(len(requests))
	results := make([]IpmiResponse, len(requests))

	a := time.Now()
	for i, r := range requests {
		go func(i int, r IpmiRequest) {
			defer wg.Done()
			al.mutex.Lock()
			defer al.mutex.Unlock()

			results[i] = fillStruct(r.Data, al, host, true)

		}(i, r)
	}
	wg.Wait()
	b := time.Now()
	c := (b.Second() - a.Second())
	log.Debug("[COLLECTION] Collection took: ", c)

	return results, nil

}

func (al *LinuxOutOfBand) ExecRaw(request IpmiRequest, host string) (*IpmiResponse, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	results := make([]IpmiResponse, 1)

	a := time.Now()
	go func(r IpmiRequest) {
		defer wg.Done()
		al.mutex.Lock()
		defer al.mutex.Unlock()

		results[0] = fillStruct(r.Data, al, host, false)

	}(request)
	wg.Wait()
	b := time.Now()
	c := (b.Second() - a.Second())
	log.Debug("[COLLECTION] Collection took: ", c)

	return &results[0], nil

}

func fillStruct(request []byte, strct *LinuxOutOfBand, addr string, isBridged bool) IpmiResponse {
	var res IpmiResponse
	res.Data = ExecIpmiToolRemote(request, strct, addr, isBridged)
	res.IsValid = 1
	return res
}

// GetPlatformCapabilities returns host capabilities
func (al *LinuxOutOfBand) GetPlatformCapabilities(requests []RequestDescription, host []string) map[string][]RequestDescription {
	validRequests := make(map[string][]RequestDescription, 0)
	var wg sync.WaitGroup

	a := time.Now()
	for _, addr := range host {
		validRequests[addr] = make([]RequestDescription, 0)
		wg.Add(len(requests))

		for _, req := range requests {
			go func(req RequestDescription, addr string) {
				al.mutex.Lock()
				defer al.mutex.Unlock()
				defer wg.Done()

				var a []byte
				if al.Protocol == "node_manager" {
					a = ExecIpmiToolRemote(req.Request.Data, al, addr, true)
				} else {
					a = ExecIpmiToolRemote(req.Request.Data, al, addr, false)
				}

				j := 0

				for i := range a {
					if a[i] == 0 {
						j++
					}
				}
				if j != len(a) {
					validRequests[addr] = append(validRequests[addr], req)
				}

			}(req, addr)
		}

		if al.Protocol == "dcmi" {
			wg.Add(1)
			// check thermal capability
			go func(req []byte, addr string) {
				al.mutex.Lock()
				defer al.mutex.Lock()
				defer wg.Done()
				a := ExecIpmiToolRemote(CmdDCMIThermalCap, al, addr, false)

				if len(a) > 5 {
					data := a[1:]
					//sdrId := data[4] + uint16(data[5])<<8
					CmdSDR[4] = data[3]
					CmdSDR[5] = data[4]
					resp := ExecIpmiToolRemote(CmdSDR, al, addr, false)

					if len(resp) > 10 {
						data = resp[1:]
						DcmiThermal.Request.Data[2] = data[9]
						validRequests[addr] = append(validRequests[addr], DcmiThermal)
					}
				}

			}(CmdDCMIThermalCap, addr)
		}

		wg.Wait()
		b := time.Now()
		c := (b.Second() - a.Second())
		log.Debug("[INIT] Initialization took: ", c)
	}

	return validRequests

}
