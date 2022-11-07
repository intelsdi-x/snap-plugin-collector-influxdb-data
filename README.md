DISCONTINUATION OF PROJECT. 

This project will no longer be maintained by Intel.

This project has been identified as having known security escapes.

Intel has ceased development and contributions including, but not limited to, maintenance, bug fixes, new releases, or updates, to this project.  

Intel no longer accepts patches to this project.
<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


    Copyright 2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**



# Snap plugin collector - InfluxDB data

Snap plugin intended to receive data previously saved in InfluxDB.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Configurable options](#configurable-options)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

### Installation
#### Download the plugin binary:
You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-influxdb-data/releases) page.
Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-influxdb-data

Clone repo into `$GOPATH/src/github/intelsdi-x/`:
```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-influxdb-data
```
Build the plugin by running make in repo:
```
$ make
```
This builds the plugin in `./build`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap#getting-started)

## Documentation
The intention of this plugin is to receive data previously saved in InfluxDB.

### Collected Metrics
This plugin has the ability to gather the following metrics:

Namespace | Description
----------|-----------------------
/intel/influxdb-data/[series_name]/[data_type]/data|Single metric received from InfluxDB
- `series_name` - name of series, namespace separator is replaced with `_`
- `data_type` - received from configuration provided by user

#Configurable options
The plugin can be configured by following parameters in config section:
- `host` - InfluxDB host (with port number)
- `database` - InfluxDB database,
- `user` - InfluxDB user,
- `password` - 'InfluxDB' password,
- `data_type` - indicates which column from response is used as a data for metric, this parameter is added for metric namespace on 4th position,
- `query` - indicates query which is used to receive data from InfluxDB.

Notice: Special characters in `query` need to be escaped.

### Examples

This is an example running snap-plugin-collector-influxdb-data to received previously saved metrics from [snap-plugin-collector-cpu](https://github.com/intelsdi-x/snap-plugin-collector-cpu)
and writing data to a file.
It is assumed that you are using the latest Snap binary and plugins.

In one terminal window, open the Snap daemon (n this case with logging set to 1 and trust disabled):
```
$ snapteld -l 1 -t 0
```
In another terminal window:

Load plugins:
```
$ snaptel plugin load snap-plugin-collector-influxdb-data
$ snaptel plugin load snap-plugin-publisher-file
```

Create a task manifest - see examplary task manifests in [examples/tasks](examples/tasks/) and create a task:

```
$ snaptel task create -t task.json
```

To stop task:
```
$ snaptel task stop <task_id>
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-influxdb-data/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-influxdb-data/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Katarzyna Kujawa](https://github.com/katarzyna-z)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
