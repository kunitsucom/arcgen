package config

import (
	"context"
	"encoding/json"
	"sync"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/internal/logs"
)

// Use a structure so that settings can be backed up.
//
//nolint:tagliatelle
type config struct {
	Version  bool   `json:"version"`
	Trace    bool   `json:"trace"`
	Debug    bool   `json:"debug"`
	Language string `json:"language"`
	// Golang
	GoColumnTag          string `json:"go_column_tag"`
	GoMethodNameTable    string `json:"go_method_name_table"`
	GoMethodNameColumns  string `json:"go_method_name_columns"`
	GoMethodPrefixColumn string `json:"go_method_prefix_column"`
	GoSliceTypeSuffix    string `json:"go_slice_type_suffix"`
}

//nolint:gochecknoglobals
var (
	globalConfig   *config
	globalConfigMu sync.RWMutex
)

func MustLoad(ctx context.Context) (rollback func(), remainingArgs []string) {
	rollback, remainingArgs, err := Load(ctx)
	if err != nil {
		err = errorz.Errorf("Load: %w", err)
		panic(err)
	}
	return rollback, remainingArgs
}

func Load(ctx context.Context) (rollback func(), remainingArgs []string, err error) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	backup := globalConfig

	cfg, remainingArgs, err := load(ctx)
	if err != nil {
		return nil, nil, errorz.Errorf("load: %w", err)
	}

	globalConfig = cfg

	rollback = func() {
		globalConfigMu.Lock()
		defer globalConfigMu.Unlock()
		globalConfig = backup
	}

	return rollback, remainingArgs, nil
}

const (
	_OptionVersion = "version"

	_OptionTrace = "trace"
	_EnvKeyTrace = "ARCGEN_TRACE"

	_OptionDebug = "debug"
	_EnvKeyDebug = "ARCGEN_DEBUG"

	_OptionLanguage = "lang"
	_EnvKeyLanguage = "ARCGEN_LANGUAGE"

	// Golang

	_OptionGoColumnTag = "go-column-tag"
	_EnvKeyGoColumnTag = "ARCGEN_GO_COLUMN_TAG"

	_OptionGoMethodNameTable = "go-method-name-table"
	_EnvKeyGoMethodNameTable = "ARCGEN_GO_METHOD_NAME_TABLE"

	_OptionGoMethodNameColumns = "go-method-name-columns"
	_EnvKeyGoMethodNameColumns = "ARCGEN_GO_METHOD_NAME_COLUMNS"

	_OptionGoMethodPrefixColumn = "go-method-prefix-column"
	_EnvKeyGoMethodPrefixColumn = "ARCGEN_GO_METHOD_PREFIX_COLUMN"

	_OptionGoSliceTypeSuffix = "go-slice-type-suffix"
	_EnvKeyGoSliceTypeSuffix = "ARCGEN_GO_SLICE_TYPE_SUFFIX"
)

// MEMO: Since there is a possibility of returning some kind of error in the future, the signature is made to return an error.
//
//nolint:funlen
func load(ctx context.Context) (cfg *config, remainingArgs []string, err error) { //nolint:unparam
	cmd := &cliz.Command{
		Name:        "arcgen",
		Usage:       "arcgen [OPTIONS] [FILE or DIR ...]",
		Description: "Generate methods that return information such as DB table names and column names from Go struct tags.",
		Options: []cliz.Option{
			&cliz.BoolOption{
				Name:        _OptionVersion,
				Description: "show version information and exit",
				Default:     cliz.Default(false),
			},
			&cliz.BoolOption{
				Name:        _OptionTrace,
				Environment: _EnvKeyTrace,
				Description: "trace mode enabled",
				Default:     cliz.Default(false),
			},
			&cliz.BoolOption{
				Name:        _OptionDebug,
				Environment: _EnvKeyDebug,
				Description: "debug mode",
				Default:     cliz.Default(false),
			},
			&cliz.StringOption{
				Name:        _OptionLanguage,
				Environment: _EnvKeyLanguage,
				Description: "programming language to generate DDL",
				Default:     cliz.Default("go"),
			},
			// Golang
			&cliz.StringOption{
				Name:        _OptionGoColumnTag,
				Environment: _EnvKeyGoColumnTag,
				Description: "column annotation key for Go struct tag",
				Default:     cliz.Default("db"),
			},
			&cliz.StringOption{
				Name:        _OptionGoMethodNameTable,
				Environment: _EnvKeyGoMethodNameTable,
				Description: "method name for table",
				Default:     cliz.Default("TableName"),
			},
			&cliz.StringOption{
				Name:        _OptionGoMethodNameColumns,
				Environment: _EnvKeyGoMethodNameColumns,
				Description: "method name for columns",
				Default:     cliz.Default("ColumnNames"),
			},
			&cliz.StringOption{
				Name:        _OptionGoMethodPrefixColumn,
				Environment: _EnvKeyGoMethodPrefixColumn,
				Description: "method prefix for column name",
				Default:     cliz.Default("ColumnName_"),
			},
			&cliz.StringOption{
				Name:        _OptionGoSliceTypeSuffix,
				Environment: _EnvKeyGoSliceTypeSuffix,
				Description: "suffix for slice type",
				Default:     cliz.Default(""),
			},
		},
	}

	remainingArgs, err = cmd.Parse(contexts.OSArgs(ctx))
	if err != nil {
		return nil, nil, errorz.Errorf("cmd.Parse: %w", err)
	}

	c := &config{
		Version:  loadVersion(ctx, cmd),
		Trace:    loadTrace(ctx, cmd),
		Debug:    loadDebug(ctx, cmd),
		Language: loadLanguage(ctx, cmd),
		// Golang
		GoColumnTag:          loadGoColumnTag(ctx, cmd),
		GoMethodNameTable:    loadGoMethodNameTable(ctx, cmd),
		GoMethodNameColumns:  loadGoMethodNameColumns(ctx, cmd),
		GoMethodPrefixColumn: loadGoMethodPrefixColumn(ctx, cmd),
		GoSliceTypeSuffix:    loadGoSliceTypeSuffix(ctx, cmd),
	}

	if c.Debug {
		logs.Debug = logs.NewDebug()
		logs.Trace.Print("debug mode enabled")
	}
	if c.Trace {
		logs.Trace = logs.NewTrace()
		logs.Debug = logs.NewDebug()
		logs.Debug.Print("trace mode enabled")
	}

	if err := json.NewEncoder(logs.Debug).Encode(c); err != nil {
		logs.Debug.Printf("config: %#v", c)
	}

	return c, remainingArgs, nil
}
