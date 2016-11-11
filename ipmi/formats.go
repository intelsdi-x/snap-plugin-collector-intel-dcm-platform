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
	"errors"
	"fmt"
)

// GenericValidator performs basic response validation. Checks response code ensures response
// has non-zero length.
type GenericValidator struct {
}

// Validate method verifies responses from IPMI device before running parsers
func (gv *GenericValidator) Validate(response IpmiResponse) error {
	if response.IsValid == 1 {
		if len(response.Data) > 0 {
			if response.Data[0] == 0 {
				return nil
			}
			return fmt.Errorf("Unexpected error code : %d", response.Data[0])
		}
		return errors.New("Zero length response")
	}
	fmt.Println("Response is not valid")
	return nil
}

// ParserCUPS extracts data from CUPS specific response format.
// Data contains info about cpu utilization and memory & io bandwidth.
type ParserCUPS struct {
	*GenericValidator
}

func GetUint16FromByteArray(data []byte, offset uint) uint16 {
	return (uint16(data[offset+1])<<8 | uint16(data[offset]))
}

// Instance of ParserCUPS
var FormatCUPS = &ParserCUPS{}

// GetMetrics method returns metric for CUPS parser: "cpu_cstate", "memory_bandwith", "io_bandwith"
func (p *ParserCUPS) GetMetrics() []string {
	return []string{"cpu_bandwith", "memory_bandwith", "io_bandwith"}
}

// Parse method returns data in human readable format
func (p *ParserCUPS) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Parsing is based on command Get CUPS Data (65h). Bytes 5:6 contains CPU CUPS dynamic load factor
	// Bytes 7:8 contains memory CUPS dynamic load factor
	// Bytes 9:10 contains IO CUPS dynamic load factor
	var names = map[string]uint{
		"cpu_bandwith":    4,
		"memory_bandwith": 6,
		"io_bandwith":     8,
	}
	for metricName, startIndex := range names {
		if response.IsValid == 1 {
			m[metricName] = GetUint16FromByteArray(response.Data, startIndex)
		} else {
			m[metricName] = 0xFFFF
		}
	}
	return m
}

// ParserCUPSIndex extracts CUPS Index from Node Manager
type ParserCUPSIndex struct {
	*GenericValidator
}

// Instance of ParserCUPS
var FormatCUPSIndex = &ParserCUPSIndex{}

// GetMetrics method returns metric for CUPS parser: "index"
func (p *ParserCUPSIndex) GetMetrics() []string {
	return []string{"index"}
}

// Parse method returns data in human readable format
func (p *ParserCUPSIndex) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	m["index"] = GetUint16FromByteArray(response.Data, 4)
	return m
}

// ParserNodeManager extracts data from Node manager response format.
// Data contains current, min, max and average value.
type ParserNodeManager struct {
	*GenericValidator
}

// Instance of ParserNodeManager
var FormatNodeManager = &ParserNodeManager{}

// GetMetrics method returns metric for CUPS parser: "current_value", "min", "max", "avg"
func (p *ParserNodeManager) GetMetrics() []string {
	return []string{"cur", "min", "max", "avg"}
}

// Parse method returns data in human readable format
func (p *ParserNodeManager) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Parsing is based on command Get Node Manager Statistics (C8h). Bytes 5:6 contains current value
	// Bytes 7:8 contains minimum value
	// Bytes 9:10 contains maximum value
	// Bytes 11:12 contains average value
	var names = map[string]uint{
		"cur": 4,
		"min": 6,
		"max": 8,
		"avg": 10,
	}
	for metricName, startIndex := range names {
		if response.IsValid == 1 {
			m[metricName] = GetUint16FromByteArray(response.Data, startIndex)
		} else {
			m[metricName] = 0xFFFF
		}
	}
	return m
}

// ParserTemp extracts temperature data.
// Data contains info about temperatures for first 4 cpus
// and 64 dimms.
type ParserTemp struct {
	*GenericValidator
}

// Instance of ParserTempMargin.
var FormatTemp = &ParserTemp{}

// GetMetrics method returns metric for temperature parser: temperature of each cpu (up to 4),
// temperature of each dimm (up to 64)
func (p *ParserTemp) GetMetrics() []string {
	a := []string{"cpu/0", "cpu/1", "cpu/2", "cpu/3"}
	for i := 0; i < 64; i++ {
		c := fmt.Sprintf("memory/dimm/%d", i)
		a = append(a, c)
	}
	return a
}

// Parse method returns data in human readable format
func (p *ParserTemp) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Parsing is based on Get CPU and Memory Temperature (4Bh). Bytes 5:8 contains temperatures of each socket (up to 4)
	if response.IsValid == 1 {
		m["cpu/0"] = uint16(response.Data[4])
		m["cpu/1"] = uint16(response.Data[5])
		m["cpu/2"] = uint16(response.Data[6])
		m["cpu/3"] = uint16(response.Data[7])
		// Bytes 9:72 contains temperatures of each dimm (up to 64)
		for i := 8; i < len(response.Data); i++ {
			a := fmt.Sprintf("memory/dimm/%d", i-8)
			m[a] = uint16(response.Data[i])
		}
	} else {
		m["cpu/0"] = 0xFFFF
		m["cpu/1"] = 0xFFFF
		m["cpu/2"] = 0xFFFF
		m["cpu/3"] = 0xFFFF
		for i := 8; i < len(response.Data); i++ {
			a := fmt.Sprintf("memory/dimm/%d", i-8)
			m[a] = 0xFFFF
		}
	}
	return m
}

// ParserPECI extracts temperature margin datas from PECI response.
// Main metric value is TJ max.
// margin_offset current value of margin offset, which is value
// of TJ max reduction.
type ParserPECI struct {
	*GenericValidator
}

// Instance of ParserPECI.
var FormatPECI = &ParserPECI{}

// GetMetrics method returns metrics for PECI parser: TJmax, margin_offset
func (p *ParserPECI) GetMetrics() []string {
	return []string{"", "margin_offset"}
}

// Parse method returns data in human readable format
func (p *ParserPECI) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Based on Send raw PECI command (40h). Byte 7 returns margin offset
	// Bytes 8:9 returns TJmax
	if response.IsValid == 1 {
		m["margin_offset"] = uint16(response.Data[6])
		m[""] = GetUint16FromByteArray(response.Data, 7)
	} else {
		m["margin_offset"] = 0xFFFF
		m[""] = 0xFFFF
	}
	return m
}

// ParserPMBus extracts temperatures of voltage regulators.
type ParserPMBus struct {
	*GenericValidator
}

// Instance of ParserPMBus.
var FormatPMBus = &ParserPMBus{}

// GetMetrics returns metrics for PMBus parser: VR[0:5]
func (p *ParserPMBus) GetMetrics() []string {
	return []string{"VR0", "VR1", "VR2", "VR3", "VR4", "VR5"}
}

// Parse method returns data in human readable format
func (p *ParserPMBus) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	var names = map[string]uint{"VR0": 4, "VR1": 6, "VR2": 8, "VR3": 10, "VR4": 12, "VR5": 14}
	//if not all VRs are present on platform return 0xFFFE
	if len(response.Data) < 14 {
		for metricName := range names {
			m[metricName] = 0xFFFE
		}
		return m
	}
	// Based on Send Raw PMBus Command (D9h). Bytes 9:N contains data received from PSU
	for metricName, startIndex := range names {
		if response.IsValid == 1 {
			m[metricName] = GetUint16FromByteArray(response.Data, startIndex)
		} else {
			m[metricName] = 0xFFFF
		}
	}
	return m
}

// ParserPSU extracts temperatures of PSU.
type ParserPSU struct {
	*GenericValidator
}

// Instance of ParserPSU.
var FormatPSU = &ParserPSU{}

// GetMetrics returns metrics for PSU Parser
func (p *ParserPSU) GetMetrics() []string {
	return []string{"0", "1"}
}

// Parse method returns data in human readable format
func (p *ParserPSU) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	m["0"] = GetUint16FromByteArray(response.Data, 4)
	m["1"] = GetUint16FromByteArray(response.Data, 6)
	return m
}

// ParserSR extracts sensor value from response to Get Sensor Record.
type ParserSR struct {
	*GenericValidator
}

// Instance of ParserSR.
var FormatSR = &ParserSR{}

// GetMetrics returns metrics for sensor reading parser: current value
func (p *ParserSR) GetMetrics() []string {
	return []string{""}
}

// Parse method returns data in human readable format
func (p *ParserSR) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Based on Get Sensor Reading (2Dh)
	m[""] = uint16(response.Data[1])
	return m
}

// ParserPolicy extracts sensor value from response to Get Power Policy.
type ParserPolicy struct {
	*GenericValidator
}

// Instance of Power Policy parser
var FormatPolicy = &ParserPolicy{}

// GetMetrics returns metrics for power limit
func (p *ParserPolicy) GetMetrics() []string {
	return []string{"power_limit"}
}

// Parse method returns data in human readable format
func (p *ParserPolicy) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	m["power_limit"] = GetUint16FromByteArray(response.Data, 13)
	return m
}

type ParserDCMIPower struct {
	*GenericValidator
}

var FormatDCMIPower = &ParserDCMIPower{}

func (p *ParserDCMIPower) GetMetrics() []string {
	return []string{"cur", "min", "max", "avg"}
}

func (p *ParserDCMIPower) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	// Parsing is based on command Get DCMI Power Reading. Bytes 3:4 contains current value
	// Bytes 5:6 contains minimum value
	// Bytes 7:8 contains maximum value
	// Bytes 9:10 contains average value
	var names = map[string]uint{
		"cur": 2,
		"min": 4,
		"max": 6,
		"avg": 8,
	}
	for metricName, startIndex := range names {
		if response.IsValid == 1 {
			m[metricName] = GetUint16FromByteArray(response.Data, startIndex)
		} else {
			m[metricName] = 0xFFFF
		}
	}
	return m
}

type ParserSensor struct {
	*GenericValidator
}

var FormatSensorReading = &ParserSensor{}

func (p *ParserSensor) GetMetrics() []string {
	return []string{"cur"}
}

func (p *ParserSensor) Parse(response IpmiResponse) map[string]uint16 {
	m := map[string]uint16{}
	m["cur"] = uint16(response.Data[1])

	return m
}
