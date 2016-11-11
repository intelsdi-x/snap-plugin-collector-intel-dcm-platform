// +build unit

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

// Tests for ipmi commands parser

package ipmi

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidator(t *testing.T) {
	Convey("Check validator", t, func() {
		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}, 1}
		a := &GenericValidator{}
		err := errors.New("Zero length response")
		validator := a.Validate(validResponse)
		So(validator, ShouldEqual, nil)
		validResponse = IpmiResponse{[]byte{}, 1}
		validator = a.Validate(validResponse)
		So(validator.Error(), ShouldEqual, err.Error())
		validResponse = IpmiResponse{[]byte{0x88, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}, 1}
		err = fmt.Errorf("Unexpected error code : %d", validResponse.Data[0])
		validator = a.Validate(validResponse)
		So(validator.Error(), ShouldEqual, err.Error())
	})
}

func TestCUPSParsing(t *testing.T) {
	Convey("Check CUPS parser", t, func() {
		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, 0x64, 0x00, 0x50, 0x00, 0x00, 0x01}, 1}
		a := &ParserCUPS{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"cpu_bandwith", "memory_bandwith", "io_bandwith"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut["cpu_bandwith"], ShouldEqual, 100)
		So(parserOut["memory_bandwith"], ShouldEqual, 80)
		So(parserOut["io_bandwith"], ShouldEqual, 256)
	})
}

func TestNodeManagerParsing(t *testing.T) {
	Convey("Check NodeManager parser", t, func() {

		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, 0x69, 0x00, 0x03, 0x00, 0x7d, 0x01, 0x6E, 0x00, 0xC7, 0x3F, 0x05, 0x56, 0xB9, 0xAD, 0x0C, 0x00, 0x50}, 1}
		a := &ParserNodeManager{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"", "min", "max", "avg"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut[""], ShouldEqual, 105)
		So(parserOut["min"], ShouldEqual, 3)
		So(parserOut["max"], ShouldEqual, 381)
		So(parserOut["avg"], ShouldEqual, 110)
	})
}

func TestPECIParsing(t *testing.T) {
	Convey("Check PECI parser", t, func() {
		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, 0x40, 0x00, 0x0A, 0x59, 0x00}, 1}
		a := &ParserPECI{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"", "margin_offset"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut[""], ShouldEqual, 89)
		So(parserOut["margin_offset"], ShouldEqual, 10)
	})
}

func TestPMBusParsing(t *testing.T) {
	Convey("Check PMBus parser", t, func() {
		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, 0x25, 0x00, 0x2A, 0x00, 0x1F, 0x00, 0x21, 0x00, 0x20, 0x00, 0x1F, 0x00}, 1}
		a := &ParserPMBus{}
		metrics := a.GetMetrics()
		parserOut := a.Parse(validResponse)
		expects := []string{"VR0", "VR1", "VR2", "VR3", "VR4", "VR5"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(expects); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		So(parserOut["VR0"], ShouldEqual, 37)
		So(parserOut["VR1"], ShouldEqual, 42)
		So(parserOut["VR2"], ShouldEqual, 31)
		So(parserOut["VR3"], ShouldEqual, 33)
		So(parserOut["VR4"], ShouldEqual, 32)
		So(parserOut["VR5"], ShouldEqual, 31)
	})
}

func TestTemperatureParsing(t *testing.T) {
	Convey("Check Temperature parser", t, func() {
		validResponse := IpmiResponse{[]byte{0x00, 0x57, 0x01, 0x00, //response header
			0x23, 0x25, 0xFF, 0xFF, //CPUS
			0xFF, 0xFF, 0x1E, 0x20, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x1F, 0x23, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x20, 0x22, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0x1E, 0x21, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 1}
		a := &ParserTemp{}
		metrics := a.GetMetrics()
		fmt.Println(metrics)
		parserOut := a.Parse(validResponse)
		expects := []string{"cpu/0", "cpu/1", "cpu/2", "cpu/3",
			"memory/dimm/0", "memory/dimm/1", "memory/dimm/2", "memory/dimm/3", "memory/dimm/4",
			"memory/dimm/5", "memory/dimm/6", "memory/dimm/7", "memory/dimm/8", "memory/dimm/9",
			"memory/dimm/10", "memory/dimm/11", "memory/dimm/12", "memory/dimm/13", "memory/dimm/14",
			"memory/dimm/15", "memory/dimm/16", "memory/dimm/17", "memory/dimm/18", "memory/dimm/19",
			"memory/dimm/20", "memory/dimm/21", "memory/dimm/22", "memory/dimm/23", "memory/dimm/24",
			"memory/dimm/25", "memory/dimm/26", "memory/dimm/27", "memory/dimm/28", "memory/dimm/29",
			"memory/dimm/30", "memory/dimm/31", "memory/dimm/32", "memory/dimm/33", "memory/dimm/34",
			"memory/dimm/35", "memory/dimm/36", "memory/dimm/37", "memory/dimm/38", "memory/dimm/39",
			"memory/dimm/40", "memory/dimm/41", "memory/dimm/42", "memory/dimm/43", "memory/dimm/44",
			"memory/dimm/45", "memory/dimm/46", "memory/dimm/47", "memory/dimm/48", "memory/dimm/49",
			"memory/dimm/50", "memory/dimm/51", "memory/dimm/52", "memory/dimm/53", "memory/dimm/54",
			"memory/dimm/55", "memory/dimm/56", "memory/dimm/57", "memory/dimm/58", "memory/dimm/59",
			"memory/dimm/60", "memory/dimm/61", "memory/dimm/62", "memory/dimm/63"}
		So(len(metrics), ShouldEqual, len(expects))
		for i := 0; i < len(metrics); i++ {
			So(metrics[i], ShouldEqual, expects[i])
		}
		for i := 0; i < len(metrics); i++ {
			So(parserOut[metrics[i]], ShouldEqual, validResponse.Data[i+4])
		}
	})
}
