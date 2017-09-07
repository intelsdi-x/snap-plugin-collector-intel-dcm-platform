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

// GenericVendor Generic list of supported requests.
var GenericVendor = []RequestDescription{
	{IpmiRequest{[]byte{0x2e, 0x65, 0x57, 0x01, 0x00, 0x02}, 6, 0x2c}, "cups", FormatCUPS},
	{IpmiRequest{[]byte{0x2e, 0x65, 0x57, 0x01, 0x00, 0x01}, 6, 0x2c}, "cups", FormatCUPSIndex},

	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x01, 0x00, 0x00}, 6, 0x2c}, "power/system", FormatNodeManager},
	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x01, 0x01, 0x00}, 6, 0x2c}, "power/cpu", FormatNodeManager},
	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x01, 0x02, 0x00}, 6, 0x2c}, "power/memory", FormatNodeManager},

	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x02, 0x00, 0x00}, 6, 0x2c}, "thermal/inlet", FormatNodeManager},
	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x05, 0x00, 0x00}, 6, 0x2c}, "thermal/outlet", FormatNodeManager},
	{IpmiRequest{[]byte{0x2e, 0x4b, 0x57, 0x01, 0x00, 0x03, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 6, 0x2c}, "thermal", FormatTemp},

	{IpmiRequest{[]byte{0x2e, 0xc8, 0x57, 0x01, 0x00, 0x04, 0x00, 0x00}, 6, 0x2c}, "airflow", FormatNodeManager},

	{IpmiRequest{[]byte{0x2e, 0x40, 0x57, 0x01, 0x00, 0x30, 0x05, 0x05, 0xa1, 0x00, 0x10, 0x00, 0x00}, 6, 0x2c}, "margin/cpu/tj", FormatPECI},

	{IpmiRequest{[]byte{0x04, 0x2d, 0x08}, 6, 0x2c}, "thermal/chipset", FormatSR},

	{IpmiRequest{[]byte{0x2e, 0xc2, 0x57, 0x01, 0x00, 0x0, 0x1}, 6, 0x2c}, "power/policy", FormatPolicy},
}

var DCMIVendor = []RequestDescription{
	{IpmiRequest{[]byte{0x2c, 0x02, 0xdc, 0x01, 0x00, 0x00}, 0, 0}, "power/system", FormatDCMIPower},
}
var DcmiThermal = RequestDescription{IpmiRequest{[]byte{0x4, 0x2d, 0x00}, 0, 0},
	"thermal/inlet", FormatSensorReading}

var CmdDCMIThermalCap = []byte{0x2c, 0x7, 0xdc, 0x01, 0x40, 0x0, 0x0}
var CmdReserverSDR = []byte{0xa, 0x22}
var CmdSDR = []byte{0xa, 0x23, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08}

var InventoryMetrics = []string{
	"inventory/firmware_version",
	"inventory/bmc_mac",
	"inventory/product_manufacturer",
	"inventory/product_name",
	"inventory/product_serial",
}

var HealthMetrics = []string{
	"health/temperature",
	"health/voltage",
	"health/fan",
	"health/processor",
	"health/powersupply",
	"health/memory",
	"health/storage",
	"health/battery"}
