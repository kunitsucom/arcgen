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
	Dialect  string `json:"dialect"`
	Language string `json:"language"`
	// Golang
	GoColumnTag          string `json:"go_column_tag"`
	GoMethodNameTable    string `json:"go_method_name_table"`
	GoMethodNameColumns  string `json:"go_method_name_columns"`
	GoMethodPrefixColumn string `json:"go_method_prefix_column"`
	GoPKTag              string `json:"go_pk_tag"`
	GoORMPackagePath     string `json:"go_orm_package_path"`
	GoORMPackageName     string `json:"go_orm_package_name"`
	GoORMTypeName        string `json:"go_orm_type_name"`
	GoORMStructName      string `json:"go_orm_struct_name"`
	GoHasManyTag         string `json:"go_has_many_tag"`
	GoHasOneTag          string `json:"go_has_one_tag"`
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
	_OptionVersion                   = "version"
	_OptionTrace, _EnvKeyTrace       = "trace", "ARCGEN_TRACE"
	_OptionDebug, _EnvKeyDebug       = "debug", "ARCGEN_DEBUG"
	_OptionDialect, _EnvKeyDialect   = "dialect", "ARCGEN_DIALECT"
	_OptionLanguage, _EnvKeyLanguage = "lang", "ARCGEN_LANGUAGE"

	//
	// Golang
	//
	_OptionGoColumnTag, _EnvKeyGoColumnTag                   = "go-column-tag", "ARCGEN_GO_COLUMN_TAG"
	_OptionGoMethodNameTable, _EnvKeyGoMethodNameTable       = "go-method-name-table", "ARCGEN_GO_METHOD_NAME_TABLE"
	_OptionGoMethodNameColumns, _EnvKeyGoMethodNameColumns   = "go-method-name-columns", "ARCGEN_GO_METHOD_NAME_COLUMNS"
	_OptionGoMethodPrefixColumn, _EnvKeyGoMethodPrefixColumn = "go-method-prefix-column", "ARCGEN_GO_METHOD_PREFIX_COLUMN"
	_OptionGoPKTag, _EnvKeyGoPKTag                           = "go-pk-tag", "ARCGEN_GO_PK_TAG"
	_OptionGoSliceTypeSuffix, _EnvKeyGoSliceTypeSuffix       = "go-slice-type-suffix", "ARCGEN_GO_SLICE_TYPE_SUFFIX"
	_OptionGoORMPackagePath, _EnvKeyGoORMPackagePath         = "go-orm-package-path", "ARCGEN_GO_ORM_PACKAGE_PATH"
	_OptionGoORMPackageName, _EnvKeyGoORMPackageName         = "go-orm-package-name", "ARCGEN_GO_ORM_PACKAGE_NAME"
	_OptionGoORMTypeName, _EnvKeyGoORMTypeName               = "go-orm-type-name", "ARCGEN_GO_ORM_TYPE_NAME"
	_OptionGoORMStructName, _EnvKeyGoORMStructName           = "go-orm-struct-name", "ARCGEN_GO_ORM_STRUCT_NAME"
	_OptionGoHasManyTag, _EnvKeyGoHasManyTag                 = "go-has-many-tag", "ARCGEN_GO_HAS_MANY_TAG"
	_OptionGoHasOneTag, _EnvKeyGoHasOneTag                   = "go-has-one-tag", "ARCGEN_GO_HAS_ONE_TAG"
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
				Name: _OptionVersion, Description: "show version information and exit",
				Default: cliz.Default(false),
			},
			&cliz.BoolOption{
				Name: _OptionTrace, Environment: _EnvKeyTrace, Description: "trace mode enabled",
				Default: cliz.Default(false),
			},
			&cliz.BoolOption{
				Name: _OptionDebug, Environment: _EnvKeyDebug, Description: "debug mode",
				Default: cliz.Default(false),
			},
			&cliz.StringOption{
				Name: _OptionDialect, Environment: _EnvKeyDialect, Description: "dialect for DML",
				Default: cliz.Default("postgres"),
			},
			&cliz.StringOption{
				Name: _OptionLanguage, Environment: _EnvKeyLanguage, Description: "programming language to generate DDL",
				Default: cliz.Default("go"),
			},
			//
			// Golang
			//
			&cliz.StringOption{
				Name: _OptionGoColumnTag, Environment: _EnvKeyGoColumnTag, Description: "column annotation key for Go struct tag",
				Default: cliz.Default("db"),
			},
			&cliz.StringOption{
				Name: _OptionGoMethodNameTable, Environment: _EnvKeyGoMethodNameTable, Description: "method name for table",
				Default: cliz.Default("TableName"),
			},
			&cliz.StringOption{
				Name: _OptionGoMethodNameColumns, Environment: _EnvKeyGoMethodNameColumns, Description: "method name for columns",
				Default: cliz.Default("ColumnNames"),
			},
			&cliz.StringOption{
				Name: _OptionGoMethodPrefixColumn, Environment: _EnvKeyGoMethodPrefixColumn, Description: "method prefix for column name",
				Default: cliz.Default("ColumnName_"),
			},
			&cliz.StringOption{
				Name: _OptionGoPKTag, Environment: _EnvKeyGoPKTag, Description: "primary key annotation key for Go struct tag",
				Default: cliz.Default("pk"),
			},
			&cliz.StringOption{
				Name: _OptionGoSliceTypeSuffix, Environment: _EnvKeyGoSliceTypeSuffix, Description: "suffix for slice type",
				Default: cliz.Default("Slice"),
			},
			&cliz.StringOption{
				Name: _OptionGoORMPackagePath, Environment: _EnvKeyGoORMPackagePath, Description: "package path for ORM",
				Default: cliz.Default(""),
			},
			&cliz.StringOption{
				Name: _OptionGoORMPackageName, Environment: _EnvKeyGoORMPackageName, Description: "package name for ORM",
				Default: cliz.Default(""),
			},
			&cliz.StringOption{
				Name: _OptionGoORMTypeName, Environment: _EnvKeyGoORMTypeName, Description: "interface type name for ORM",
				Default: cliz.Default("ORM"),
			},
			&cliz.StringOption{
				Name: _OptionGoORMStructName, Environment: _EnvKeyGoORMStructName, Description: "struct name for ORM",
				Default: cliz.Default(""),
			},
			&cliz.StringOption{
				Name: _OptionGoHasOneTag, Environment: _EnvKeyGoHasOneTag, Description: "\"hasOne\" annotation key for Go struct tag",
				Default: cliz.Default("hasOne"),
			},
			&cliz.StringOption{
				Name: _OptionGoHasManyTag, Environment: _EnvKeyGoHasManyTag, Description: "\"hasMany\" annotation key for Go struct tag",
				Default: cliz.Default("hasMany"),
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
		Dialect:  loadDialect(ctx, cmd),
		Language: loadLanguage(ctx, cmd),
		// Golang
		GoColumnTag:          loadGoColumnTag(ctx, cmd),
		GoMethodNameTable:    loadGoMethodNameTable(ctx, cmd),
		GoMethodNameColumns:  loadGoMethodNameColumns(ctx, cmd),
		GoMethodPrefixColumn: loadGoMethodPrefixColumn(ctx, cmd),
		GoPKTag:              loadGoPKTag(ctx, cmd),
		GoSliceTypeSuffix:    loadGoSliceTypeSuffix(ctx, cmd),
		GoORMPackagePath:     loadGoORMPackagePath(ctx, cmd),
		GoORMPackageName:     loadGoORMPackageName(ctx, cmd),
		GoORMTypeName:        loadGoORMTypeName(ctx, cmd),
		GoORMStructName:      loadGoORMStructName(ctx, cmd),
		GoHasOneTag:          loadGoHasOneTag(ctx, cmd),
		GoHasManyTag:         loadGoHasManyTag(ctx, cmd),
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
