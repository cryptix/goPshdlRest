# goPSHDLwpsync


A [PSHDL](http://pshdl.org) REST-API client written in go.

Uploads changes from the local copy of a workspace to the remote one.

## Installation
With [go](http://golang.org) installed:
```
go get github.com/cryptix/goPSHDLwpsync
```

_I am working on precompiled binaries, too. If you don't want to program in go yourself you don't have to install it._

## Usage

Currently there are two commands. __open__ and __new__.


### Open
This downloads an existing workspace to your harddrive and starts watching it for changes.
_It creates a new directory in your current working directory with the Id of the workspace as a name._

```
goPSHDLwpsync open <wid>
```

### New
This Requests a new workspace on the API. It also starts watching the supplied path for changes.


```
goPSHDLwpsync new <path>
```


## TODO/Ideas

* Add a Watch command to fetch compiled simulation files.
* Supply Problems and Errors to editors (for [SublimeLinter](https://github.com/SublimeLinter/SublimeLinter) for instance)
* Add command line options to set Name and Email for new workspaces