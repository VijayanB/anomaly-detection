![Test Workflow](https://github.com/VijayanB/esad/workflows/Build%20and%20Test%20Anomaly%20detection%20commandline%20tool/badge.svg)

# Open Distro for Elasticsearch AD CLI

The AD CLI component in Open Distro for Elasticsearch (ODFE) is a command line interface for ODFE AD plugin.
his CLI provides greater flexibility of use. User can use CLI to easily do things that are difficult or sometimes impossible to do with kibana UI. This doesn’t use any additional  system resources to load any of graphical part, thus making it simpler and faster than UI. 

It only supports [Open Distro for Elasticsearch (ODFE) AD Plugin](https://opendistro.github.io/for-elasticsearch-docs/docs/ad/)
You must have the ODFE AD plugin installed to your Elasticsearch instance to connect. 
Users can run this CLI from MacOS and Linux, and connect to any valid Elasticsearch end-point such as Amazon Elasticsearch Service (AES).The ESAD CLI implements AD APIs.

## Features

* Create Detectors
* Start, Stop, Delete Detectors
* Create named profiles to connect to ES cluster

## Install

Launch your local Elasticsearch instance and make sure you have the Open Distro for Elasticsearch AD plugin installed.

To install the AD CLI:


1. Install from source:

    ```
    $ go get github.com/VijayanB/esad/
    ```

## Configure

Before using the AWS CLI, you need to configure your AWS credentials. You can do this in several ways:

* Configuration command
* Config file

The quickest way to get started is to run the `esad profile create`

```
$ esad profile create
Enter profile's name: dev
ES Anomaly Detection Endpoint: https://localhost:9200
ES Anomaly Detection User: admin
ES Anomaly Detection Password:
```

To use a config file, create a YAML file like this
```
profiles:
- endpoint: https://localhost:9200
  username: admin
  password: foobar
  name: default
- endpoint: https://odfe-node1:9200
  username: admin
  password: foobar
  name: dev
```
and place it on ~/.esad/config.yaml. if you wish to place the shared credentials file in a different location than the one specified above, you need to tell aws-cli where to find it. Do this by setting the appropriate environment variable:

```
export ESAD_CONFIG_FILE=/path/to/config_file
```
You can have multiple profiles defined in the configuration file. You can then specify which profile to use by using the --profile option. If no profile is specified the `default` profile is used.



## Basic Commands

AN ESAD CLI has following structure
```
$ esad <command> <subcommand> [flags and parameters]
```
For example to start detector:
```
$ esad start [detector-name-pattern]
```
To view help documentation, use one of the following:
```
$ esad --help
$ esad <command> --help
$ esad <command> <subcommand> --help
```
To get the version of the ESAD CLI:
```
$ esad --version
```

## Getting Help

The best way to interact with our team is through GitHub. You can open an [issue](https://github.com/opendistro-for-elasticsearch/anomaly-detection/issues) and tag accordingly.

## Copyright

Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
