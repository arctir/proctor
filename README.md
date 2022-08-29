# Proctor

A CLI for understanding the relationships of processes on a host.

## Install and Usage

## How it works

Under the hood, proctor leverages our library plib, which is meant to be an operating-system agnostic way to expose details of the processes on a given host. Surfacing this information, proctor has functionality to create an overview of what's known about a host's state. As an initial use-case, plib introspects procfs on Linux hosts. Over time, this will expand to:

* Windows
* FreeBSD
* Non-procfs Linux introspection

## Why Proctor

Proctor is the command-line representation for core functionality used in Arctir's platform.
