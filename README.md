# co2mini
Golang Module to use co2meter code derived from https://hackaday.io/project/5301-reverse-engineering-a-low-cost-usb-co-monitor

## Requirements

Ubuntu

```
apt install libhidapi-dev libudev-dev golang
```

macOS

```
brew install hidapi golang
```

## Build

```
go build
```

## Usage

```
./co2mini [-i <interval>]
```

## Example

```
{"time":1641003315,"co2":579,"temp":20.912500000000023}
{"time":1641003320,"co2":579,"temp":20.912500000000023}
{"time":1641003325,"co2":579,"temp":20.850000000000023}
{"time":1641003330,"co2":577,"temp":20.912500000000023}
{"time":1641003335,"co2":577,"temp":20.912500000000023}
```
