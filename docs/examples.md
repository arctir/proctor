# Examples

This page contains usage examples for proctor.

## CLI examples

Below are example of interacting with proctor as a CLI.

### Source examples

These examples relate to the `proctor source ...` command.

#### Retrieve differences in commits between 2 tags

Diffing 2 tags is achieved by running `commits diff ...`, and specifying tags
with `--tag1` and `--tag2`.

```sh
$ proctor source commits diff https://github.com/nixos/nix --tag1 2.1.3 --tag2 2.1.2                                                       [10:03:11]
```

Results in:

```txt
+-----------------------+
| commits only in 2.1.3 |
+-----------------------+
+------------------------------------------+--------------------------------+---------------------+------------------+
|                   SHA                    |            MESSAGE             |       AUTHOR        |    TIMESTAMP     |
+------------------------------------------+--------------------------------+---------------------+------------------+
| 9011a85ef6e8a3f99f34b6fe7c89c15852ea1697 | nix-profile-daemon: remove cru | mjbauer95@gmail.com | 2018-10-01 22:51 |
| 1c99c1ba43acaa24264c6cbcd01c7095c70042fc | Update docs to describe how s3 | graham@grahamc.com  | 2018-10-01 22:49 |
| 96e03a39ec05d5b46aabe87e547e5c4215dc137a | Ensure download thread livenes | edolstra@gmail.com  | 2018-10-01 22:49 |
| 51d11e9e0dccde508a36e753f8a17ba666cddcd1 | sinkToSource(): Start the coro | edolstra@gmail.com  | 2018-10-01 22:48 |
| 87ad88f28de741e0af648eb13af09206ab37e5fd | Make NAR header check more rob | edolstra@gmail.com  | 2018-10-01 22:48 |
| 373cc12d63b34c93aaf15c448cf057efd99f5608 | Bump version                   | edolstra@gmail.com  | 2018-10-01 22:47 |
+------------------------------------------+--------------------------------+---------------------+------------------+
+-----------------------+
| commits only in 2.1.2 |
+-----------------------+
+-----+---------+--------+-----------+
| SHA | MESSAGE | AUTHOR | TIMESTAMP |
+-----+---------+--------+-----------+
+-----+---------+--------+-----------+
```

#### List available releases tied to a repo

> Note: this currently only works with GitHub

```sh
proctor source artifacts list https://github.com/kubernetes-sigs/cluster-api
```

Results in:

```txt
+---------------+---------------+-----------+
|      TAG      |     TITLE     | ARTIFACTS |
+---------------+---------------+-----------+
| v1.3.1        | v1.3.1        |        15 |
| v1.2.8        | v1.2.8        |        15 |
| v1.3.0        | v1.3.0        |        15 |
| v1.2.7        | v1.2.7        |        15 |
| v1.3.0-rc.1   | v1.3.0-rc.1   |        15 |
| v1.2.6        | v1.2.6        |        15 |
| v1.3.0-rc.0   | v1.3.0-rc.0   |        15 |
| v1.3.0-beta.1 | v1.3.0-beta.1 |        15 |
| v1.2.5        | v1.2.5        |        15 |
| v1.3.0-beta.0 | v1.3.0-beta.0 |        15 |
| v1.2.4        | v1.2.4        |        15 |
| v1.2.3        | v1.2.3        |        15 |
| v1.2.2        | v1.2.2        |        15 |
| v1.1.6        | v1.1.6        |        13 |
| v1.2.1        | v1.2.1        |        15 |
| v1.2.0        | v1.2.0        |        15 |
| v1.2.0-rc.0   | v1.2.0-rc.0   |        15 |
| v1.2.0-beta.2 | v1.2.0-beta.2 |        15 |
| v1.1.5        | v1.1.5        |        13 |
| v1.2.0-beta.1 | v1.2.0-beta.1 |        14 |
| v1.2.0-beta.0 | v1.2.0-beta.0 |        14 |
| v1.1.4        | v1.1.4        |        13 |
| v1.1.3        | v1.1.3        |        14 |
| v1.0.5        | v1.0.5        |        11 |
| v0.4.8        | v0.4.8        |        11 |
| v1.1.2        | v1.1.2        |        14 |
| v1.1.1        | v1.1.1        |        14 |
| v1.1.0        | v1.1.0        |        14 |
| v1.0.4        | v1.0.4        |        11 |
| v0.4.7        | v0.4.7        |        11 |
+---------------+---------------+-----------+
```

#### Retrieve artifacts associated with a release

> Note: this currently only works with GitHub

`artifacts get` is used to retrieve artifacts associated with a release. The
associated tag must be specified with the `--tag` flag.

```sh
proctor source artifacts get https://github.com/kubernetes-sigs/cluster-api --tag v1.2.5
```

Results in:


```txt
+--------------------------------------------+--------------------------+-----------------------------------------------------------------------------------+
|                   ASSSET                   |       CONTENT-TYPE       |                                        URL                                        |
+--------------------------------------------+--------------------------+-----------------------------------------------------------------------------------+
| bootstrap-components.yaml                  | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896664 |
| cluster-api-components.yaml                | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896669 |
| cluster-template-development.yaml          | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896677 |
| clusterclass-quick-start.yaml              | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896667 |
| clusterctl-darwin-amd64                    | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896668 |
| clusterctl-darwin-arm64                    | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896673 |
| clusterctl-linux-amd64                     | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896670 |
| clusterctl-linux-arm64                     | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896671 |
| clusterctl-linux-ppc64le                   | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896665 |
| clusterctl-windows-amd64.exe               | application/octet-stream | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896666 |
| control-plane-components.yaml              | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896663 |
| core-components.yaml                       | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896676 |
| infrastructure-components-development.yaml | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896672 |
| metadata.yaml                              | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896675 |
| runtime-sdk-openapi.yaml                   | text/yaml                | https://api.github.com/repos/kubernetes-sigs/cluster-api/releases/assets/83896674 |
+--------------------------------------------+--------------------------+-----------------------------------------------------------------------------------+
```

#### Retrieve all commits in a tag

`commits list` is used to get all commits in a tag. The `--tag` flag is used to
specify the flag.

```sh
proctor source commits ls https://github.com/kubernetes-sigs/cluster-api --tag v1.2.5
```

Results in:

```txt
+------------------------------------------+--------------------------------+--------------------------------+
|                   SHA                    |            MESSAGE             |             AUTHOR             |
+------------------------------------------+--------------------------------+--------------------------------+

< -- snipped -- >

| cd8d6ae832b47ca3e3960e305bc7cc09bd81bca9 | Fix issues mentioned in #10520 | zml@google.com                 |
| 95f8c2d6c093e91764c3c9823717ca3086c0acd8 | wait until a token shows up to | dbsmith@google.com             |
| 70bee7b6060571e3caf8c4bfef82d794989d7926 | Updating heapster version to v | vishnuk@google.com             |
| 1e507fe243d954b2861792fd6ce76c968dfd544a | Enable InfluxDB/Grafana for GC | saadali@google.com             |
| 232c76de4c5cdf20786edd4315dd428c187831d6 | Updating heapster version to v | vishnuk@google.com             |
| 6e8e1d877edbd52b946d1dae85d571219448d470 | e2e test for addon update      | biskup@google.com              |
| d7c72f79ddd3fb0681a85e0acab2a7c74db689db | kube_addons - Adding variable  | jeffreyrobertbean@gmail.com    |
| 8e425f508411ffbdc30ae909691135702845946f | Distribute the cluster CA cert | robertbailey@google.com        |
| 534a8d6f72b74e99a02c30efacb8175e32567d5b | kube-addon-update.sh           | biskup@google.com              |
| 8fd21e459fdfe89352cc1951f5027960895a24e6 | Enable InfluxDB/Grafana for GC | saadali@google.com             |
| 06d3aac90e2222808d4f07720cd027bf6ff9bda0 | Create LimitRange object for c | dawnchen@google.com            |
| 107cbc92a5f3da326acd0570a1d349629aa965ec | Updating /cluster to use v1bet | krousey@google.com             |
| 6a907b1135278f9eaef8983fda754002fedfa54c | Enable Google Cloud Monitoring | saadali@google.com             |
| 8f7090bce12194909ad3126bdd2fb52480838730 | Make copyright ownership state | eparis@redhat.com              |
| 68f9aafe0f7ca0f7e7b12937d3d45c5a18b43c70 | kube2sky using kubeconfig secr | cjcullen@google.com            |
| 1d533e43a38321fcb0a9844b5a5f3e028fd4d998 | Convert Elasticsearch logging  | satnam@google.com              |
| 8a926f2aba842368e8783ecec0028cc4ed23c9ee | Create system secrets in kubec | etune@google.com               |
| ffe3ee4472d66cd04a80a829ab0bd811efe76943 | Fix kube-addon retrying.       | etune@google.com               |
| 2ca3917fcb7f31de2a5ad790ad2a597ad85df913 | Make secrets at cluster startu | etune@google.com               |
| 93e34c6eedeaecc5664e1695dd2bc14a1529392c | Use same addons script for ini | etune@google.com               |
| 30944057bcb5b975242615a6b3748f9ac373c2d7 | Retry kube-addons creation if  | abshah@google.com              |
| f6caab55f860a0e0e14d03cfb72f4ca70866a351 | Retry object creation with --v | zml@google.com                 |
| e6a85857210134303f8eeb0d51032a8ba03be806 | Missing boilerplate            | decarr@redhat.com              |
| ca6a76a5ce88bb5937c0eac7f1395c3e937dedb8 | Various vagrant fixes, etcd 2. | decarr@redhat.com              |
| e4119f09125cf314ec4ac685ccd94b3d3f0370a2 | Deferred creation of SkyDNS, m | zml@google.com                 |
| 6ad30a0711466107edb3e91c51f50355cdb49c8f | final commit                   | mikedanese@google.com          |
+------------------------------------------+--------------------------------+--------------------------------+
```

#### Retrieve all authors from a repository

`commits list` with the `--authors` flag can be used to list al contributors
along their quantity of commits.

```sh
proctor source commits ls https://github.com/kubernetes-sigs/cluster-api --tag v1.2.5 --authors
```

```txt
+---------+----------------------------+--------------------------------------------------------------+
| COMMITS |            NAME            |                            EMAIL                             |
+---------+----------------------------+--------------------------------------------------------------+
|    2894 | k8s-ci-robot               | k8s-ci-robot@users.noreply.github.com                        |
|     553 | Vince Prignano             | vincepri@vmware.com                                          |
|     397 | fabriziopandini            | fpandini@vmware.com                                          |
|     275 | Stefan Bueringer           | buringerst@vmware.com                                        |
|     258 | Justin Santa Barbara       | justin@fathomdb.com                                          |
|     148 | killianmuldoon             | kmuldoon@vmware.com                                          |
|     126 | Chuck Ha                   | chuckh@vmware.com                                            |
|     108 | Jacob Beacham              | beacham@google.com                                           |
|      86 | Jason DeTiberus            | detiberusj@vmware.com                                        |
|      86 | Feng Min                   | fmin@google.com                                              |
|      77 | Cecile Robert-Michon       | cerobert@microsoft.com                                       |

<-- snipped -->
```

### Process examples

#### List all processes known to host

> ⚠️: By default, proctor caches the process table after your first request. To
> query the process tree and reset the cache, use the --reset-cache flag.

`list` can be used to retrieve all processes on the system. Note that you may
wish to use `sudo` since list will only returns processes you have privileges to
access.

```sh
proctor process ls
```

Results in:

```txt
+--------+-------------------------------------+-----------------------------------------------------------+------------------------------------------------------------------+
|  PID   |                NAME                 |                         LOCATION                          |                               SHA                                |
+--------+-------------------------------------+-----------------------------------------------------------+------------------------------------------------------------------+
|  67228 | gvfsd-dnssd                         | /usr/lib/gvfsd-dnssd                                      | b41d9070813a414b276c3918ca0bf10fde2f3d19a4f3f4fea3e73022e4b8640d |
|  43225 | bash                                | /usr/bin/bash                                             | 864925e8e16b3c2bc999c77e4959f20b4834e48f49b966e550e00e13dc01f9b7 |
|   1334 | xfwm4                               | /usr/bin/xfwm4                                            | 024da2825e5d28dcdf73987c36bd28dba183e012f24af61b803e224e6cab7a69 |
|  43266 | chromium                            | /usr/lib/chromium/chromium                                | 9e49ab8e46367c7229f07a7227753360e39932f276c523a0ddff58fcbbbb0080 |
|   1499 | gvfsd-metadata                      | /usr/lib/gvfsd-metadata                                   | 32e765023a42946a8edd3fb577f2f8f3cff5c1903b4699f73ad25cc7994f3a49 |
|   1412 | pasystray                           | /usr/bin/pasystray                                        | b1615a5001e07a7f03e0a742b134c72c5d3b3d32fbe1b0f6cd7ef2cc5fc4aa0e |
|  99654 | nvim                                | /usr/bin/nvim                                             | 1299d93ce9940efea02904690bc2d1712ab8a1ae24c4d005f569572e5e8390bd |
| 102507 | proctor                             | /home/josh/proctor/out/proctor                            | 08230f6667b5a76baeb9885178a3b3186d3cc0755a19bc27ab08366b8819a767 |
|  99814 | chromium                            | /usr/lib/chromium/chromium                                | 9e49ab8e46367c7229f07a7227753360e39932f276c523a0ddff58fcbbbb0080 |
|   1614 | obexd                               | /usr/lib/bluetooth/obexd                                  | bb7263de1c23507e0aaaef6346a32e696107da32ae2c941a1870efb9c6821ec4 |
|  43242 | chromium                            | /usr/lib/chromium/chromium                                | 9e49ab8e46367c7229f07a7227753360e39932f276c523a0ddff58fcbbbb0080 |

<-- snipped -->
```

#### Retrieve process and all its relative processes

> ⚠️: By default, proctor caches the process table after your first request. To
> query the process tree and reset the cache, use the --reset-cache flag.

`tree` can be used to get a process's information along with all parent
processes.

```sh
sudo proctor process tree 354446
```

Results in:

```txt
+--------+---------+--------------------------+------------------------------------------------------------------+
|  PID   |  NAME   |         LOCATION         |                               SHA                                |
+--------+---------+--------------------------+------------------------------------------------------------------+
| 354446 | dockerd | /usr/bin/dockerd         | 90291368a1c4d217fe4609eafe04c088359cbb62da5a795f07c99c0afdb02d80 |
|      1 | systemd | /usr/lib/systemd/systemd | 1408f42d86aff4db2807fd9fe69d13d6d4b5943fbddff0ad09380adec6abea21 |
+--------+---------+--------------------------+------------------------------------------------------------------+
```

#### Retrieve a fingerprint for a process and its relatives

> ⚠️: By default, proctor caches the process table after your first request. To
> query the process tree and reset the cache, use the --reset-cache flag.

`finger-print` is used to create a unique finger-print for a process. Today,
this is calculated by taking the SHA value of the binary and all its parents
processes and then creating another unique hash by creating a hash of those
hashes.

```sh
proctor process finger-print 354446
```

Results in:

```txt
653d0e436631b3c25a62876756625ec27b7748ef236b0da8423542a4401bdff0
```

#### Retrieve detailed output for a processes

By default, `process` prints in table format with limited information. To get
all the details around the process, you can output using `--output json`. Below
is an example of getting the JSON output for any processes named `dockerd`.

```sh
proctor process get --name dockerd --output json
```

Results in:

```json
{
  "354446": {
    "ID": 354446,
    "BinarySHA": "90291368a1c4d217fe4609eafe04c088359cbb62da5a795f07c99c0afdb02d80",
    "CommandName": "dockerd",
    "CommandPath": "/usr/bin/dockerd",
    "FlagsAndArgs": "",
    "ParentProcess": 1,
    "IsKernel": false,
    "HasPermission": true,
    "Type": "",
    "OSSpecific": {
      "ID": 354446,
      "FileName": "(dockerd)",
      "State": "S",
      "ParentID": 1,
      "ProcessGroup": 354446,
      "SessionID": 354446,
      "TTY": 0,
      "TTYProcessGroup": -1,
      "TaskFlags": "1077936384",
      "MinorFaultQuantity": 3171,
      "MinorFaultWithChildQuantity": 8409,
      "MajorFaultQuantity": 599,
      "MajorFaultWithChildQuantity": 101,
      "UserModeTime": 7,
      "KernalTime": 5,
      "UserModeTimeWithChild": 4,
      "KernalTimeWithChild": 3,
      "Priority": 20,
      "Nice": 0,
      "ThreadQuantity": 22,
      "ItRealValue": 0,
      "StartTime": 23252534,
      "VirtualMemSize": 2398375936,
      "ResidentSetMemSize": 19614,
      "RSSByteLimit": 9223372036854776000,
      "StartCode": "0x55d8cc0c3000",
      "EndCode": "0x55d8cd8fb921",
      "StartStack": "0x7ffe9778fc00",
      "ExtendedStackPointerAddress": 0,
      "ExtendedInstructionPointer": 0,
      "SignalPendingQuantity": 0,
      "SignalsBlockedQuantity": 1002055680,
      "SignalsIgnoredQuantity": 0,
      "SiganlsCaughtQuantity": 2143420159,
      "PlaceHolder1": 0,
      "PlaceHolder2": 0,
      "PlaceHolder3": 0,
      "ExitSignal": 17,
      "CPU": 0,
      "RealtimePriority": 0,
      "SchedulingPolicy": 0,
      "TimeSpentOnBlockIO": 0,
      "GuestTime": 0,
      "GuestTimeWithChild": 0,
      "StartDataAddress": "0x55d8cdcea030",
      "EndDataAddress": "0x55d8cf3b8058",
      "HeapExpansionAddress": "0x55d8d05af000",
      "StartCMDAddress": "0x7ffe9778fed1",
      "EndCMDAddress": "0x7ffe9778feeb",
      "StartEnvAddress": "0x7ffe9778feeb",
      "EndEnvAddress": "0x7ffe9778ffe7",
      "ExitCode": 0
    }
  }
}
```

## Library usage

Below are example of using proctor as a library in your Go projects.

( coming soon! )
