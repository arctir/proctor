# Proctor

A CLI for introspecting what's running on a host.

## Install and Usage

1. Install proctor.

	```sh
	go install github.com/arctir/proctor
	```

	> While this repo is private, additionally steps may be required to run the
	> above command. Alternatively, you can `git clone` the repository and run
	> `make install` to achieve the same.

1. Run a command. Examples below.

	* `proctor ls`: list all processes.
	* `proctor tree [process-name]`: list a process's relationship.
	* `proctor get -o json [process-name]`: get details on a process.
	* `proctor fp [process-name]`: Returns a checksum based on a process's relationships.
		> By default, it uses the hashes of its and all parent process binaries.
	* `proctor help`: list commands.

## Why Proctor

Proctor is a command-line utility that surfaces some of the functionality in
Arctir's core libraries. These libraries are used in building our commercial
platforms. We leverage it to prove out concepts at a low-level, however, you
may find benefit in using proctor to introspect and understand your software at
a runtime level.

## How it works

Under the hood, proctor leverages our library plib, which evolving into a
operating-system agnostic way to expose details of the processes on a given
introspects procfs on Linux hosts. Over time, this will expand to:

* Windows
* FreeBSD
* Non-procfs Linux introspection

Architecturally, the package structure is as follows.

![package-architecture](https://user-images.githubusercontent.com/6200057/187974676-0652bad7-0d89-4450-8327-6d48304bf709.png)

## Development

Run `make help` to see development tasks.
