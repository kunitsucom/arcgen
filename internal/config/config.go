package config

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"

	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/internal/logs"
)

// Use a structure so that settings can be backed up.
//
//nolint:tagliatelle
type config struct {
	Version   bool   `json:"version"`
	Trace     bool   `json:"trace"`
	Debug     bool   `json:"debug"`
	Timestamp string `json:"timestamp"`
	Language  string `json:"language"`
	Source    string `json:"source"`
	// Golang
	ColumnTagGo        string `json:"column_tag_go"`
	MethodPrefixColumn string `json:"column_method_prefix"`
	MethodPrefixGlobal string `json:"global_method_prefix"`
}

//nolint:gochecknoglobals
var (
	globalConfig   *config
	globalConfigMu sync.RWMutex
)

func MustLoad(ctx context.Context) (rollback func()) {
	rollback, err := Load(ctx)
	if err != nil {
		err = errorz.Errorf("Load: %w", err)
		panic(err)
	}
	return rollback
}

func Load(ctx context.Context) (rollback func(), err error) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	backup := globalConfig

	cfg, err := load(ctx)
	if err != nil {
		return nil, errorz.Errorf("load: %w", err)
	}

	globalConfig = cfg

	rollback = func() {
		globalConfigMu.Lock()
		defer globalConfigMu.Unlock()
		globalConfig = backup
	}

	return rollback, nil
}

const (
	_OptionVersion = "version"

	_OptionTrace = "trace"
	_EnvKeyTrace = "ARCGEN_TRACE"

	_OptionDebug = "debug"
	_EnvKeyDebug = "ARCGEN_DEBUG"

	_OptionTimestamp = "timestamp"
	_EnvKeyTimestamp = "ARCGEN_TIMESTAMP"

	_OptionLanguage = "lang"
	_EnvKeyLanguage = "ARCGEN_LANGUAGE"

	_OptionSource = "src"
	_EnvKeySource = "ARCGEN_SOURCE"

	_OptionDestination = "dst"
	_EnvKeyDestination = "ARCGEN_DESTINATION"

	// Golang

	_OptionColumnTagGo = "column-tag-go"
	_EnvKeyColumnTagGo = "ARCGEN_COLUMN_TAG_GO"

	_OptionMethodPrefixGlobal = "method-prefix-global"
	_EnvKeyMethodPrefixGlobal = "ARCGEN_METHOD_PREFIX_GLOBAL"

	_OptionMethodPrefixColumn = "method-prefix-column"
	_EnvKeyMethodPrefixColumn = "ARCGEN_METHOD_PREFIX_COLUMN"
)

// MEMO: Since there is a possibility of returning some kind of error in the future, the signature is made to return an error.
//
//nolint:funlen
func load(ctx context.Context) (cfg *config, err error) { //nolint:unparam
	cmd := &cliz.Command{
		Name:        "ARCgen",
		Description: "Generate DDL from annotated source code.",
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
				Name:        _OptionTimestamp,
				Environment: _EnvKeyTimestamp,
				Description: "code generation timestamp",
				Default:     cliz.Default(time.Now().Format(time.RFC3339)),
			},
			&cliz.StringOption{
				Name:        _OptionLanguage,
				Environment: _EnvKeyLanguage,
				Description: "programming language to generate DDL",
				Default:     cliz.Default("go"),
			},
			&cliz.StringOption{
				Name:        _OptionSource,
				Environment: _EnvKeySource,
				Description: "source file or directory",
				Default:     cliz.Default("/dev/stdin"),
			},
			&cliz.StringOption{
				Name:        _OptionDestination,
				Environment: _EnvKeyDestination,
				Description: "destination file or directory",
				Default:     cliz.Default("/dev/stdout"),
			},
			// Golang
			&cliz.StringOption{
				Name:        _OptionColumnTagGo,
				Environment: _EnvKeyColumnTagGo,
				Description: "column annotation key for Go struct tag",
				Default:     cliz.Default("db"),
			},
			&cliz.StringOption{
				Name:        _OptionMethodPrefixGlobal,
				Environment: _EnvKeyMethodPrefixGlobal,
				Description: "global method prefix",
				Default:     cliz.Default("Get"),
			},
			&cliz.StringOption{
				Name:        _OptionMethodPrefixColumn,
				Environment: _EnvKeyMethodPrefixColumn,
				Description: "column method prefix",
				Default:     cliz.Default("ColumnName_"),
			},
		},
	}

	if _, err := cmd.Parse(contexts.Args(ctx)); err != nil {
		return nil, errorz.Errorf("cmd.Parse: %w", err)
	}

	c := &config{
		Version:   loadVersion(ctx, cmd),
		Trace:     loadTrace(ctx, cmd),
		Debug:     loadDebug(ctx, cmd),
		Timestamp: loadTimestamp(ctx, cmd),
		Language:  loadLanguage(ctx, cmd),
		Source:    loadSource(ctx, cmd),
		// Golang
		ColumnTagGo:        loadColumnTagGo(ctx, cmd),
		MethodPrefixGlobal: loadMethodPrefixGlobal(ctx, cmd),
		MethodPrefixColumn: loadMethodPrefixColumn(ctx, cmd),
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

	return c, nil
}
