# [arcgen](https://github.com/kunitsucom/arcgen)

[![license](https://img.shields.io/github/license/kunitsucom/arcgen)](LICENSE)
[![pkg](https://pkg.go.dev/badge/github.com/kunitsucom/arcgen)](https://pkg.go.dev/github.com/kunitsucom/arcgen)
[![goreportcard](https://goreportcard.com/badge/github.com/kunitsucom/arcgen)](https://goreportcard.com/report/github.com/kunitsucom/arcgen)
[![workflow](https://github.com/kunitsucom/arcgen/workflows/go-lint/badge.svg)](https://github.com/kunitsucom/arcgen/tree/main)
[![workflow](https://github.com/kunitsucom/arcgen/workflows/go-test/badge.svg)](https://github.com/kunitsucom/arcgen/tree/main)
[![workflow](https://github.com/kunitsucom/arcgen/workflows/go-vuln/badge.svg)](https://github.com/kunitsucom/arcgen/tree/main)
[![codecov](https://codecov.io/gh/kunitsucom/arcgen/graph/badge.svg?token=Y19kZ7UtVZ)](https://codecov.io/gh/kunitsucom/arcgen)
[![sourcegraph](https://sourcegraph.com/github.com/kunitsucom/arcgen/-/badge.svg)](https://sourcegraph.com/github.com/kunitsucom/arcgen)

## Overview

`arcgen` is a tool that generates methods that return information such as DB table names and column names from Go struct tags.

## Example

```console
$ # == 1. Prepare your annotated model source code ================================
$ cat <<"EOF" > /tmp/sample.go
package sample

// User is a user model struct.
//
// db: table: Users
type User struct {
    ID   int64  `db:"Id"   spanddl:"STRING(36)  NOT NULL"`
    Name string `db:"Name" spanddl:"STRING(255) NOT NULL"`
    Age  int64  `db:"Age"  spanddl:"INT64       NOT NULL"`
}

// Group is a group model struct.
//
type Group struct {
    ID          int64  `db:"Id"          spanddl:"STRING(36)   NOT NULL"`
    Name        string `db:"Name"        spanddl:"STRING(255)  NOT NULL"`
    Description string `db:"Description" spanddl:"STRING(2048) NOT NULL"`
}
EOF

$ # == 2. generate file ================================
$ arcgen --src /tmp/sample.go
INFO: 2023/11/12 03:56:59 arcgen.go:33: source: /tmp/sample.go

$ # == 3. Check generated file ================================
$ cat /tmp/sample.db.gen.go
// Code generated by arcgen. DO NOT EDIT.
//
// source: tmp/sample.go

package sample

func (s *User) TableName() string {
    return "Users"
}

func (s *User) ColumnNames() []string {
    return []string{"Id", "Name", "Age"}
}

func (s *User) ColumnName_ID() string {
    return "Id"
}

func (s *User) ColumnName_Name() string {
    return "Name"
}

func (s *User) ColumnName_Age() string {
    return "Age"
}

func (s *Group) ColumnNames() []string {
    return []string{"Id", "Name", "Description"}
}

func (s *Group) ColumnName_ID() string {
    return "Id"
}

func (s *Group) ColumnName_Name() string {
    return "Name"
}

func (s *Group) ColumnName_Description() string {
    return "Description"
}
```

## Installation

### pre-built binary

```bash
VERSION=v0.0.3

# download
curl -fLROSs https://github.com/kunitsucom/arcgen/releases/download/${VERSION}/arcgen_${VERSION}_darwin_arm64.zip

# unzip
unzip -j arcgen_${VERSION}_darwin_arm64.zip '*/arcgen'
```

### go install

```bash
go install github.com/kunitsucom/arcgen/cmd/arcgen@latest
```

## Usage

```console
$ arcgen --help
Usage:
    arcgen [options]

Description:
    Generate methods that return information such as DB table names and column names from Go struct tags.

options:
    --version (default: false)
        show version information and exit
    --trace (env: ARCGEN_TRACE, default: false)
        trace mode enabled
    --debug (env: ARCGEN_DEBUG, default: false)
        debug mode
    --lang (env: ARCGEN_LANGUAGE, default: go)
        programming language to generate DDL
    --src (env: ARCGEN_SOURCE, default: /dev/stdin)
        source file or directory
    --column-tag-go (env: ARCGEN_COLUMN_TAG_GO, default: db)
        column annotation key for Go struct tag
    --method-name-table (env: ARCGEN_METHOD_NAME_TABLE, default: TableName)
        method name for table
    --method-name-columns (env: ARCGEN_METHOD_NAME_COLUMNS, default: ColumnNames)
        method name for columns
    --method-prefix-column (env: ARCGEN_METHOD_PREFIX_COLUMN, default: ColumnName_)
        method prefix for column name
    --help (default: false)
        show usage
```

## TODO

- lang
  - [x] Support `go`
