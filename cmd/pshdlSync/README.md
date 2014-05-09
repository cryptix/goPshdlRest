# pshdlSync


A [PSHDL](http://pshdl.org) REST-API client written in go.

Uploads changes from the local copy of a workspace to the remote one.

## Installation
With [go](http://golang.org) installed:
```
go get github.com/cryptix/goPshdlRest/cmd/pshdlSync
```

_I'll add binarys soon. [goxc](https://github.com/laher/goxc) is aweosme_
## Usage

This looks for a `.wid` file in your current directory.
If it finds one, it looks for 16 characters inside, specifing the workspace id.

If the workspace exists it downloads all pshdl code and starts watching for changes.

If there is no `.wid` file it will create a new one.


## TODO
* write ID of a newly created workspace to `.wid`
