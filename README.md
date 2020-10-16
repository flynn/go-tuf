# go-tuf
[![build](https://github.com/theupdateframework/go-tuf/workflows/build/badge.svg)](https://github.com/theupdateframework/go-tuf/actions?query=workflow%3Abuild) [![Coverage Status](https://coveralls.io/repos/github/theupdateframework/go-tuf/badge.svg)](https://coveralls.io/github/theupdateframework/go-tuf) [![PkgGoDev](https://pkg.go.dev/badge/github.com/theupdateframework/go-tuf)](https://pkg.go.dev/github.com/theupdateframework/go-tuf) [![Go Report Card](https://goreportcard.com/badge/github.com/theupdateframework/go-tuf)](https://goreportcard.com/report/github.com/theupdateframework/go-tuf)  

This is a Go implementation of [The Update Framework (TUF)](http://theupdateframework.com/),
a framework for securing software update systems.

## Directory layout

A TUF repository has the following directory layout:

```
.
├── keys
├── repository
│   └── targets
└── staged
    └── targets
```

The directories contain the following files:

* `keys/` - signing keys (optionally encrypted) with filename pattern `ROLE.json`
* `repository/` - signed manifests
* `repository/targets/` - hashed target files
* `staged/` - either signed, unsigned or partially signed manifests
* `staged/targets/` - unhashed target files

## CLI

`go-tuf` provides a CLI for managing a local TUF repository.

### Install

```
go get github.com/theupdateframework/go-tuf/cmd/tuf@delegation
```

### Commands

#### `tuf init [--consistent-snapshot=false]`

Initializes a new repository.

This is only required if the repository should not generate consistent
snapshots (i.e. by passing `--consistent-snapshot=false`). If consistent
snapshots should be generated, the repository will be implicitly
initialized to do so when generating keys.

#### `tuf gen-key [--expires=<days>] <role>`

Prompts the user for an encryption passphrase (unless the
`--insecure-plaintext` flag is set), then generates a new signing key and
writes it to the relevant key file in the `keys` directory. It also stages
the addition of the new key to the `root` manifest.

#### `tuf add [<path>...]`

Hashes files in the `staged/targets` directory at the given path(s), then
updates and stages the `targets` manifest. Specifying no paths hashes all
files in the `staged/targets` directory.

#### `tuf remove [<path>...]`

Stages the removal of files with the given path(s) from the `targets` manifest
(they get removed from the filesystem when the change is committed). Specifying
no paths removes all files from the `targets` manifest.


#### `dele-gen-key [--expires=<days>] <role>`

Creates a new delegation role's key. Prompts the user for an encryption passphrase
(unless the `--insecure-plaintext` flag is set), then generates a new signing key and
writes it to the relevant key file in the `keys` directory. It also stages
the addition of the new key to the `target` manifest.

#### `dele-add <names>  [<path>...]`

Hashes files in the `staged/targets` directory at the given path(s), then
updates and stages the delegated target manifest. Specifying no paths hashes all
files in the `staged/targets` directory.

#### `dele-remove <name> [<path>...]`

Stages the removal of files with the given path(s) from the certain non-top target manifest
(they get removed from the filesystem when the change is committed). Specifying
no paths removes all files from the certain non-top target manifest.

#### `tuf snapshot [--compression=<format>]`

Expects a staged, fully signed `targets` manifest and stages an appropriate
`snapshot` manifest. It optionally compresses the staged `targets` manifest.

#### `tuf timestamp`

Stages an appropriate `timestamp` manifest. If a `snapshot` manifest is staged,
it must be fully signed.

#### `tuf sign ROLE`

Signs the given role's staged manifest with all keys present in the `keys`
directory for that role.

#### `tuf commit`

Verifies that all staged changes contain the correct information and are signed
to the correct threshold, then moves the staged files into the `repository`
directory. It also removes any target files which are not in the top-level `targets`
and non-top target manifests.

#### `tuf regenerate [--consistent-snapshot=false]`

Recreates the `targets` manifest based on the files in `repository/targets`.
This function has not been implemented yet.

#### `tuf clean`

Removes all staged manifests and targets.

#### `tuf root-keys`

Outputs a JSON serialized array of root keys to STDOUT. The resulting JSON
should be distributed to clients for performing initial updates.

#### `tuf target-keys`

Outputs a JSON serialized array of target keys to STDOUT. The resulting JSON
should be distributed to clients for performing initial updates.


For a list of supported commands, run `tuf help` from the command line.


### Examples

The following are example workflows for managing a TUF repository with the CLI.

The `tree` commands do not need to be run, but their output serve as an
illustration of what files should exist after performing certain commands.

Although only two machines are referenced (i.e. the "root" and "repo" boxes),
the workflows can be trivially extended to many signing machines by copying
staged changes and signing on each machine in turn before finally committing.

Some key IDs are truncated for illustrative purposes.

#### Create signed root manifest

Generate a root key on the root box:
##### Note that passphrase cannot be none

```
$ tuf gen-key root
Enter root keys passphrase:
Repeat root keys passphrase:
Generated root key with ID 184b133f

$ tree .
.
├── keys
│   └── root.json
├── repository
└── staged
    ├── root.json
    └── targets
```

Copy `staged/root.json` from the root box to the repo box and generate targets,
snapshot and timestamp keys:

```
$ tree .
.
├── keys
├── repository
└── staged
    ├── root.json
    └── targets

$ tuf gen-key targets
Enter targets keys passphrase:
Repeat targets keys passphrase:
Enter root keys passphrase: 
Generated targets key with ID 8cf4810c

$ tuf dele-gen-key r01
Enter r01 keys passphrase: 
Repeat r01 keys passphrase: 
Enter targets keys passphrase: 
Generated r01 key with ID 4d6ddd68

$ tuf gen-key snapshot
Enter snapshot keys passphrase:
Repeat snapshot keys passphrase:
Enter root keys passphrase:
Generated snapshot key with ID 3e070e53

$ tuf gen-key timestamp
Enter timestamp keys passphrase:
Repeat timestamp keys passphrase:
Enter root keys passphrase:
Generated timestamp key with ID a3768063

$ tree .
.
├── keys
│   ├── r01.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
```

Copy `staged/root.json` from the repo box back to the root box and sign it:

```
$ tree .
.
├── keys
│   ├── root.json
├── repository
└── staged
    ├── root.json
    └── targets

$ tuf sign root.json
Enter root keys passphrase:
```

The staged `root.json` can now be copied back to the repo box ready to be
committed alongside other manifests.

#### Add target files

Assuming a staged, signed `root` manifest and the files to add exists at
`staged/targets/foo/bar/baz.txt` and `staged/targets/sin.txt`:

```
$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
        ├── sin.txt
        └── foo
            └── bar
                └── baz.txt

$ tuf add foo/bar/baz.txt
Enter targets keys passphrase:

$tuf dele-add r01 sin.txt
Enter r01 keys passphrase:

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    ├── targets
    │   ├── sin.txt
    │   └── foo
    │       └── bar
    │           └── baz.txt
    ├── r01.json
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    ├── snapshot.json
    ├── targets
    │   ├── sin.txt
    │   └── foo
    │       └── bar
    │           └── baz.txt
    ├── r01.json
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   ├── sin.txt
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── r01.json
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Remove target files

Assuming the files to remove are `repository/targets/foo/bar/baz.txt` 
and `repository/targets/sin.txt`:

```
$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   ├── sin.txt
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf remove foo/bar/baz.txt
Enter targets keys passphrase:

$tuf dele-remove r01 sin.txt
Enter r01 keys passphrase:

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   ├── sin.txt
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   ├── sin.txt
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    ├── snapshot.json
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── r01.json
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Regenerate manifests based on targets tree

```
$ tree .
.
├── keys

│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf regenerate
Enter targets keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    ├── snapshot.json
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Update timestamp.json

```
$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Modify key thresholds

TODO

## Client

For the client package, see https://godoc.org/github.com/theupdateframework/go-tuf/client.

For the client CLI, see https://github.com/theupdateframework/go-tuf/tree/master/cmd/tuf-client.
