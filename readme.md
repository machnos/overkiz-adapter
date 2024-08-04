# Overkiz Adapter #
A simple GO library for managing Overkiz tokens and controlling Overkiz Kizconnect devices.

## About ##
Overkiz provides white-label software and devices to automate residential appliances like shutters and other 
window covers. This project aims to integrate those components to other devices that are able to execute web-hooks.
As an example you could open your Somfy shutters when a Shelly smoke alarm detects smoke.

## Installation ##
This repository contains the sources for two command-line binaries. You need a GO 1.22 (or higher) build environment
to create the binaries. From the root of the project execute the following commands to build those binaries:
```shell
go build cmd/overkiz-adapter.go
go build cmd/overkiz-token.go
```
The binaries will be available in the root of the project afterward.

## Usage ##
Both binaries are implemented as command-line tools. The `overkiz-adapter` binary exposes a http endpoint though. 

### overkiz-token ###
The `overkiz-token` binary is used to manage the tokens that need to be provisioned to your Kizconnect devices (like a
Somfy TaHoma switch). The development mode must be enabled on the device. The provisioning is a one-time step. After
completion the `overkiz-adapter` will talk to your devices directly. No cloud access is necessary afterward.
The executable will query for new devices every 5 minutes.

Assuming the development mode is enabled on your device you can execute the following steps to get a token.
```shell
./overkiz-token login --region=<region> --username=<username> --password=<password>
```

The region can be one of
* europe
* middle east
* africa 
* asia 
* pacific
* north america

The username and password are the same as you used for your product registration.

The Overkiz api needs a session cookie to authenticate, so after a successful login the cookie is stored in 
`<USER_HOME>/.machnos/overkiz-token/cookie.json`. This cookie will be used for other commands of the `overkiz-token`

Next we need to create a token by executing the following command
```shell
./overkiz-token create --region=<region> --pin=<device pin>
```

Again you need to provide any of the valid regions. The pin is visible on Kizconnect (like) device. When succeeded a 
token will be displayed on the console. Please keep this token on a secure place. It will not be possible to display it
again! This pin is necessary in the configuration of the `overkiz-adapter`.

The `overkiz-token` has some more commands which can be shown by just executing the binary without any commands:
```shell
./overkiz-token
Manages Overkiz tokens

Usage:
  overkiz-token [command] [options]

Available Commands:
  login      Login to Overkiz
  logout     Logout from Overkiz
  list       List all tokens
  create     Create a new token
  delete     Delete an existing token
```

The help of a certain command will be displayed by executing the command without any option. For example 
```shell
./overkiz-token list
List Overkiz tokens

Usage:
  overkiz-token list [options]

Required Options:
  --region   Region, one of "europe", "middle east", "africa", "asia", "pacific" or "north america"
  --pin      The PIN of the gateway
```

### overkiz-adapter ###
The `overkiz-adapter` exposes some convenient http endpoints that allow you to execute actions on certain devices. 
First we need to create a json configuration file for the executable. The content of the file needs to be something like this
```json
{
  "token": "<overkiz token>",
  "host": "gateway-<device pin>.local",
  "http": {
    "interface": "0.0.0.0",
    "port": 8080,
    "context_root": "/",
    "allowed_hosts": ["my-personal-computer", "127.0.0.1"],
    "behind_proxy": false
  }
}
```
* *token* - The token is the token you've received with the `overkiz-token create` command.
* *host* - The hostname or ip address of the gateway.
* *http.interface* The interface to listen on. 
* *http.port* The port to listen on.
* *http.context_root* The context root the api should have.
* *http.allowed_hosts* An optional list of domain names or ip addresses that are allowed to access the api.
* *behind_proxy* Set to true if the api is accessed via a proxy. The application will then look at the X-Forwarded-For header to determine if access is allowed.

Once the configuration file is created you can start the application by executing
```shell
./overkiz-adapter --config-file=<path-to-configuration-file>
```

The api currently exposes the following endpoints

| Path                                       | Function                            |
|--------------------------------------------|-------------------------------------|
| <context_root>/api/v1/devices              | List all devices                    |
| <context_root>/api/v1/devices/{class}      | List all devices of a certain class | 
| <context_root>/api/v1/RollerShutters/open  | Opens all RollerShutter devices     |
| <context_root>/api/v1/RollerShutters/close | Closes all RollerShutter devices    |

