# pshdlSync


A [PSHDL](http://pshdl.org) REST-API client written in go.

Uploads changes from the local copy of a workspace to the remote one.

## Installation
With [go](http://golang.org) installed:
```
go get github.com/cryptix/pshdlSync
```

_I am working on precompiled binaries, too. If you don't want to program in go yourself you don't have to install it._

## Usage

Currently there are three commands. __open__, __new__ and __stream__.


### Open
This downloads an existing workspace to your harddrive and starts watching it for changes.
_It creates a new directory in your current working directory with the Id of the workspace as a name._

```
pshdlSync open <wid>
```

### New
This Requests a new workspace on the API. It also starts watching the supplied path for changes.


```
pshdlSync new <path>
```

### Stream
This uses the streaming API to hook on events. The default is just to display events. You can trigger the download of generated code with the flags __-vhdl__ and __-csim__. 

```
pshdlSync stream [flags] <wid>
```

## TODO/Ideas

* Add flags to configure behaviour
* Supply Problems and Errors to editors (for [SublimeLinter](https://github.com/SublimeLinter/SublimeLinter) for instance)
* Add command line options to set Name and Email for new workspaces