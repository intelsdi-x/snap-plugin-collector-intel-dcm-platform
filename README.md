DISCONTINUATION OF PROJECT. 

This project will no longer be maintained by Intel.

This project has been identified as having known security escapes.

Intel has ceased development and contributions including, but not limited to, maintenance, bug fixes, new releases, or updates, to this project.  

Intel no longer accepts patches to this project.
# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**


# This Plugin is in Alpha.

It is being worked on in the open in an effort to replace the [node manager](https://github.com/intelsdi-x/snap-plugin-collector-node-manager) plugin. Please report any bugs by creating an [issue](https://github.com/intelsdi-x/snap-plugin-collector-intel-dcm-platform/issues/new).

# snap collector plugin - Intel OPEN DCM Platform Collector

 Plugin to collect inventory, health, power/thermal related metrics from platforms which expose such data through Intel Node Manager, DCMI or generic IPMI interfaces.
 Currently it is using Ipmitool or open ipmi driver to collect data. 
 This plugin is based on the snap-plugin-collector-node-manager and extend the coverage to more IPMI enabled platforms.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

 Plugin collects specified metrics in-band on OS level

### System Requirements

 - Support server platforms with Intel Node Manager, DCMI or generic IPMI support.
 - Currently it works only on Linux Servers (will be tested on a subset of Linux distributions)
 - Ipmitool needs to be installed on platform

### Installation
#### Download Intel OPEN DCM Platform plugin binary:
You can get the pre-built binaries for your OS and architecture at snap's [Github Releases](https://github.com/intelsdi-x/snap/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-intel-dcm-platform  
Clone repo into `$GOPATH/src/github/intelsdi-x/`:
```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-intel-dcm-platform
```
Build the plugin by running make in repo:
```
$ make
```
This builds the plugin in `/build/rootfs`

### Configuration and Usage

 On OS level user needs to load modules:
  - ipmi_msghandler
  - ipmi_devintf
  - ipmi_si
 
Those modules provides specific IPMI device which can collect data from NM, DCMI or generic IPMI

There are currently 7 configuration options:
 - mode - defines mode of plugin work, possible values: legacy_inband, legacy_inband_openipmi, oob
 - channel - defines communication channel address (default: "0x00")
 - slave - defines target address (default: "0x00")
 - user - for OOB mode only, user for authentication to remote host
 - password - for OOB mode only, password for authentication to remote host
 - host - for OOB mode only, BMC IP address of host which will be monitored OOB
 - protocol - defines the communication protocol used to collect metric data, possible values: node_manager, dcmi, ipmi


Sample configuration of intel dcm platform plugin:
```
{
    "control" : {
        "plugins": {
            "collector": {
                "intel-dcm-platform": {
                    "all": {
                        "protocol": "node_manager",
                        "mode": "legacy_inband",
                        "channel": "0x06",
                        "slave": "0x2C"
                    }
                }
            }
        }
    }
}
```

## Documentation

### Collected Metrics
This plugin has the ability to gather the following metrics:

Namespace | Data Type | Description
----------|-----------|-----------------------
/intel/dcm/airflow/cur | uint16 | Current Volumetric Airflow
/intel/dcm/airflow/avg | uint16 | Average Volumetric Airflow 
/intel/dcm/airflow/max | uint16 | Maximal Volumetric Airflow 
/intel/dcm/airflow/min | uint16 | Minimal Volumetric Airflow 
/intel/dcm/cups/cpu_cstate | uint16 | CUPS CPU Bandwidth
/intel/dcm/cups/io_bandwith | uint16 | CUPS I/O Bandwidth
/intel/dcm/cups/memory_bandwith | uint16 | CUPS Memory Bandwidth
/intel/dcm/power/cpu/cur | uint16 | Current CPU power consumption
/intel/dcm/power/cpu/avg | uint16 | Average CPU power consumption
/intel/dcm/power/cpu/max | uint16 | Maximal CPU power consumption
/intel/dcm/power/cpu/min | uint16 | Minimal CPU power consumption
/intel/dcm/power/policy/power_limit | uint16 | Power policy
/intel/dcm/margin/cpu/tj  | uint16 | Margin-to-throttle functional  (CPU)
/intel/dcm/margin/cpu/tj/margin_offset | uint16 | Margin-to-spec reliability (CPU)
/intel/dcm/power/memory/cur | uint16 | Current Memory power consumption
/intel/dcm/power/memory/avg | uint16 | Average Memory power consumption
/intel/dcm/power/memory/max | uint16 | Maximal Memory power consumption
/intel/dcm/power/memory/min | uint16 | Minimal Memory power consumption
/intel/dcm/power/system/cur | uint16 | Current Platform power consumption
/intel/dcm/power/system/avg | uint16 | Average Platform power consumption
/intel/dcm/power/system/max | uint16 | Maximal Platform power consumption
/intel/dcm/power/system/min | uint16 | Minimal Platform power consumption
/intel/dcm/temperature/cpu/cpu/<cpu_id> | uint16 | Current CPU temperature
/intel/dcm/temperature/pmbus/VR/<VR_id> | uint16 | Current VR's temperature
/intel/dcm/temperature/memory/dimm/<dimm_id> | uint16 | Current Memory dimms temperature
/intel/dcm/temperature/outlet/cur | uint16 | Current Outlet (exhaust air) temperature
/intel/dcm/temperature/outlet/avg | uint16 | Average Outlet (exhaust air) temperature
/intel/dcm/temperature/outlet/max | uint16 | Maximal Outlet (exhaust air) temperature
/intel/dcm/temperature/outlet/min | uint16 | Minimal Outlet (exhaust air) temperature
/intel/dcm/temperature/inlet/cur | uint16 | Current Inlet Temperature
/intel/dcm/temperature/inlet/avg | uint16 | Average Inlet Temperature
/intel/dcm/temperature/inlet/max | uint16 | Maximal Inlet Temperature
/intel/dcm/temperature/inlet/min | uint16 | Minimal Inlet Temperature
/intel/dcm/inventory/firmware_version | string | Version of management firmware
/intel/dcm/inventory/bmc_mac | string | MAC address string of BMC
/intel/dcm/inventory/product_manufacturer | string | Product Manufacturer name queried from FRU
/intel/dcm/inventory/product_name | string | Product Name queried from FRU
/intel/dcm/inventory/product_serial | string | Product Serial number queried from FRU
/intel/dcm/health/processor | string | "OK" for good state and other message for corresponding processor error
/intel/dcm/health/memory | string | "OK" for good state and other message for corresponding memory error
/intel/dcm/health/fan | string | "OK" for good state and other message for corresponding fan error
/intel/dcm/health/powersupply | string | "OK" for good state and other message for corresponding power supply error
/intel/dcm/health/driverslot | string | "OK" for good state and other message for corresponding driver error

### Metric Tags
Namespace | Tag | Description
----------|-----|------------
/intel/dcm/* | source | Host IP address

### Examples
In one terminal window, open the Snap daemon with path to config file (see [examplary config file](examples/configs/)). In this case with logging set to 1 and trust disabled:
```
$ snapteld -l 1 -t 0 --config config.json
```

Load snap-plugin-collector-intel-dcm-platform plugin:
```
$ snaptel plugin load snap-plugin-collector-intel-dcm-platform
```

See available metrics for your system:
```
$ snaptel metric list
```

Create task manifest to use Intel OPEN DCM Platform plugin (or use an [examplary task manifest](examples/tasks/)):
```
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "5s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/dcm/power/system/avg": {},
                "/intel/dcm/power/system/max": {},
                "/intel/dcm/power/system/min": {},
	        "/intel/dcm/inventory/product_name ": {},
	        "/intel/dcm/inventory/product_manufacturer ":{},
	        "/intel/dcm/inventory/firmware_version":{},
	        "/intel/dcm/health/powersupply":{},
	        "/intel/dcm/health/fan":{},
	        "/intel/dcm/health/processor":{},
	        "/intel/dcm/thermal/inlet/cur":{},
	        "/intel/dcm/thermal/inlet/max":{},
            },
            "config": {
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "file",
                    "plugin_version": 2,
                    "config": {
                        "file": "/tmp/published_dcminfo"
                    }
                }
            ]
        }
    }
}

```

Get file plugin for publishing and load it:
```
$ wget  http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ snaptel plugin load snap-plugin-publisher-file
```

Create a task:
```
$ snaptel task create -t dcm-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```
See realtime output from `snaptel task watch <task_id>` (CTRL+C to exit)
```
snaptel task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850

Watching Task (02dd7ff4-8106-47e9-8b86-70067cd0a850):
NAMESPACE                        DATA    TIMESTAMP
/intel/dcm/power/cpu/avg         44      2017-04-14 12:18:39.31235067 +0000 UTC
/intel/dcm/power/memory/avg      18      2017-04-14 12:18:39.31235067 +0000 UTC
/intel/dcm/power/system/avg      129     2017-04-14 12:18:39.31235067 +0000 UTC
/intel/dcm/thermal/inlet/avg     26      2017-04-14 12:18:39.31235067 +0000 UTC
```

### Roadmap
As we launch this plugin, we have a few items in mind for the next release:
- Remove IPMI tool support
- More health info support
- Scalability support for multiple hosts

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Dancy Ding](https://github.com/dancyding)
* Author: [Xin Dong](https://github.com/xindongIntel)
* Author: [Jialei Wang](https://github.com/WangJialei-A )

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
