# Proctor

A toolkit enabling deep introspection of various aspects of software, from
source code to runtime. Proctor contains the libraries used in Arctir's platform
services. Atop these libraries is a CLI that acts as a client enabling users to
easily try-out and leverage the power of these introspection tools.

Proctor may be particularly helpful to a developer, SRE, or devops humans
wishing to quickly introspect aspects of their software and stacks. For a
quickstart and few examples, see our [examples docs](docs/examples.md).

Our intent with proctor is to share libraries with the open source community,
help folks build other introspection-oriented tooling atop, and get feedback on
our ideas and approaches.

## Install

### As a CLI

At this time, there is no official package manager support for proctor. However,
you can easily install the binary with the go tool or download it from [GitHub
releases](https://github.com/arctir/proctor/releases). To install proctor with
the go tool, run:

```
go install github.com/arctir/proctor
```

This will place the proctor binary, for your target architecture, in `$GOBIN`.
If desired, move proctor to your `$PATH`.

### As a library

Within the `pkg` directory are multiple libraries. To use these libraries in
your own Go project(s), you can import them via go mod. For example, here's how
you could import plib, our process library responsible for introspection of
various operating-system's execution abstraction:

```sh
go get github.com/arctir/proctor/pkg/plib
```

Similarly, the Go doc describing this library are available at
[pkg.go.dev](https://pkg.go.dev/github.com/arctir/proctor/pkg/plib).

Once installed, run `proctor help` or visit [our documentation](TODO(joshrosso)) for details on
usage.

## Development

Run `make help` to see development tasks.

Feel free to open issues and/or pull requests.

## Proctor CLI Interaction

```
proctor process (p) # a client for plib, which exposes process details for
                      various operating systems.

                { list (ls) } # list all processes on the system. Note this will only
                return processes accesible to your user, you may need to run as
                sudo.
                { get (g) } # retrieve process(s) based on the name (via argument)
                or 
                { tree (t) } # A process and all of its parent(s) process details.
                { finger-print (fp) } # creates a unique sha256 hash based on the
                combination of multiple other hashes.
                { cache (c) } # caches all processes locally for quicker interaction
                and to ensure you can introspect from a specific point in time.
                When you don't create a cache, all operations read the entirity
                of the system's known processes.

proctor source (s) # introspect a source repository 

proctor deps (d) # provide a graph of dependencies for a source repository or
                   binary.
```
