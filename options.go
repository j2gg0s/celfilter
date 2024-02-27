package celfilter

import (
	"time"

	"github.com/google/cel-go/cel"
)

type Option func(*rawConverter) error

func WithSQLName(name, sqlName string) Option {
	return func(cvt *rawConverter) error {
		cvt.sqlNameMap[name] = sqlName
		return nil
	}
}

func WithLocation(loc *time.Location) Option {
	return func(cvt *rawConverter) error {
		cvt.timeLoc = loc
		return nil
	}
}

func WithTimeFormat(format string) Option {
	return func(cvt *rawConverter) error {
		cvt.timeFormat = format
		return nil
	}
}

func WithEnvOption(opts ...cel.EnvOption) Option {
	return func(cvt *rawConverter) error {
		cvt.envOpts = append(cvt.envOpts, opts...)
		return nil
	}
}

func WithPrefix(prefix string) Option {
	return func(cvt *rawConverter) error {
		cvt.prefix = prefix
		return nil
	}
}
