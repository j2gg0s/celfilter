package celfilter

import (
	"fmt"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/stdlib"
)

func NewEnv(opts ...cel.EnvOption) (*cel.Env, error) {
	opts = append([]cel.EnvOption{cel.EagerlyValidateDeclarations(false)}, opts...)

	env, err := getStdEnv()
	if err != nil {
		return nil, fmt.Errorf("new env: %w", err)
	}

	return env.Extend(opts...)
}

var (
	stdEnvOnce sync.Once
	stdEnv     *cel.Env
	stdEnvErr  error
)

func getStdEnv() (*cel.Env, error) {
	stdEnvOnce.Do(func() {
		lib, err := NewLib()
		if err != nil {
			stdEnvErr = err
			return
		}
		stdEnv, stdEnvErr = cel.NewCustomEnv(
			lib,
			cel.EagerlyValidateDeclarations(true),
			cel.Variable("i", cel.DynType),
		)
	})
	return stdEnv, stdEnvErr
}

var allowedFns = map[string]bool{
	operators.LogicalAnd: true,
	operators.LogicalOr:  true,
	operators.LogicalNot: true,

	operators.Equals:        true,
	operators.NotEquals:     true,
	operators.Less:          true,
	operators.LessEquals:    true,
	operators.Greater:       true,
	operators.GreaterEquals: true,

	operators.Add:      true,
	operators.Subtract: true,
	operators.Multiply: true,
	operators.Divide:   true,
	operators.Modulo:   true,
	operators.Negate:   true,

	overloads.StartsWith: true,
	overloads.EndsWith:   true,

	overloads.TypeConvertTimestamp: true,
}

type filterLib struct {
	funcs map[string]bool
}

func NewLib(opts ...func(*filterLib) error) (cel.EnvOption, error) {
	lib := filterLib{
		funcs: allowedFns,
	}
	for _, opt := range opts {
		if err := opt(&lib); err != nil {
			return nil, err
		}
	}
	return cel.Lib(lib), nil
}

func EnableFunctions(names ...string) func(*filterLib) error {
	return func(lib *filterLib) error {
		for _, name := range names {
			if !lib.funcs[name] {
				lib.funcs[name] = true
			}
		}
		return nil
	}
}

func DisableFunctions(names ...string) func(*filterLib) error {
	return func(lib *filterLib) error {
		for _, name := range names {
			if lib.funcs[name] {
				delete(lib.funcs, name)
			}
		}
		return nil
	}
}

func (filterLib) LibraryName() string {
	return "cel.j2gg0s.filter"
}

func (lib filterLib) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{}
	for _, fn := range stdlib.Functions() {
		if !lib.funcs[fn.Name()] {
			continue
		}
		fn := fn
		opts = append(opts, cel.Function(fn.Name(),
			func(*decls.FunctionDecl) (*decls.FunctionDecl, error) {
				return fn, nil
			}))
	}
	for _, typ := range stdlib.Types() {
		typ := typ
		opts = append(opts, cel.Variable(typ.Name(), typ.Type()))
	}
	opts = append(opts, cel.Macros(cel.StandardMacros...))
	return opts
}

func (filterLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
