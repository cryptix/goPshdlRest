goPshdlRest
===========

This package provides access to the [PSHDL](http://pshdl.org) REST API.


## Features
* Create new and open existing Workspaces
* Upload/Download/Delete files to Workspaces
* Get Events of Workspace changes through the StreamingService

## Clients
There are currently two clients in the `cmd` folder.

`pshdlSync` is used to push local changes to the remote api.
It's only one-way currently. Check out [localhelper](http://code.pshdl.org/pshdl.localhelper/wiki/Home) if you want two-way.


`pshdlCompilat` watches a workspace for Events and downloads generated code
Currently VHDL and C but the others would be simple to add.

## Documentation
Checkout [godoc.org](http://godoc.org/github.com/cryptix/goPshdlRest/api).
It's not 100% complete but I'm working on it.


## TODO
* turn service into resources that contain a client
* More Tests!
* More Documentation!
* Add Validate() and RequestSimCode() to clients
* Write a client that retreives pshdl errors to integrate with editors.
