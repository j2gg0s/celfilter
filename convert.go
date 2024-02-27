package celfilter

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type rawConverter struct {
	sqlNameMap map[string]string
	timeLocMap map[string]*time.Location

	timeLoc    *time.Location
	timeFormat string

	prefix string

	envOpts []cel.EnvOption
	*cel.Env
}

func NewConverter(opts ...Option) (*rawConverter, error) {
	cvt := &rawConverter{
		sqlNameMap: map[string]string{},
		timeLocMap: map[string]*time.Location{},

		timeLoc:    time.Local,
		timeFormat: "2006-01-02 15:04:05.999",

		prefix: "i",
	}
	for _, opt := range opts {
		if err := opt(cvt); err != nil {
			return nil, err
		}
	}

	env, err := NewEnv(append(cvt.envOpts, cel.Variable(cvt.prefix, types.DynType))...)
	if err != nil {
		return nil, fmt.Errorf("extend cel.Env: %w", err)
	}
	cvt.Env = env

	return cvt, nil
}

func (cvt *rawConverter) Convert(expr string) (string, error) {
	ast, iss := cvt.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return "", fmt.Errorf("parse&check cel expr %s: %w", expr, iss.Err())
	}
	return cvt.convert(ast.NativeRep().Expr())
}

func (cvt *rawConverter) convert(expr celast.Expr) (string, error) {
	switch expr.Kind() {
	case celast.CallKind:
		return cvt.convertCall(expr.AsCall())
	case celast.SelectKind:
		return cvt.convertSelect(expr.AsSelect())
	case celast.LiteralKind:
		return cvt.convertLiteral(expr.AsLiteral())
	}

	return "", fmt.Errorf("unsupport expr kind: %v", expr.Kind())
}

func (cvt *rawConverter) convertLiteral(val ref.Val) (string, error) {
	switch val.Type().TypeName() {
	case types.StringType.TypeName():
		return fmt.Sprintf(`"%s"`, val.Value().(string)), nil
	}
	val = val.ConvertToType(types.StringType)
	return val.Value().(string), nil
}

func (cvt *rawConverter) convertSelect(expr celast.SelectExpr) (string, error) {
	name := expr.FieldName()

	op := expr.Operand()
	for op.Kind() == celast.SelectKind {
		name += "." + op.AsSelect().FieldName()
		op = op.AsSelect().Operand()
	}

	if op.Kind() != celast.IdentKind {
		return "", fmt.Errorf("except ident: %v", op.Kind())
	}
	if op.AsIdent() != cvt.prefix {
		return "", fmt.Errorf("except %s, got %s", cvt.prefix, op.AsIdent())
	}

	sqlName, ok := cvt.sqlNameMap[name]
	if ok {
		return sqlName, nil
	}
	return name, nil
}

func (cvt *rawConverter) convertCall(expr celast.CallExpr) (string, error) {
	name := expr.FunctionName()
	target := expr.Target()
	args := expr.Args()

	switch name {
	case operators.LogicalAnd, operators.LogicalOr,
		operators.Equals, operators.NotEquals,
		operators.Less, operators.LessEquals,
		operators.Greater, operators.GreaterEquals,
		operators.Add, operators.Subtract,
		operators.Multiply, operators.Divide, operators.Modulo:

		op, _ := findReverse(name)
		arg0, err := cvt.convert(args[0])
		if err != nil {
			return "", err
		}
		arg1, err := cvt.convert(args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", arg0, op, arg1), nil

	case operators.LogicalNot,
		operators.Negate:

		op, _ := findReverse(name)

		arg0, err := cvt.convert(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s)", op, arg0), nil

	case overloads.TypeConvertTimestamp:
		if args[0].Kind() != celast.LiteralKind {
			return "", fmt.Errorf("except literal in %s: %v", name, args[0].Kind())
		}

		arg := args[0].AsLiteral()
		switch arg.Type().TypeName() {
		case types.StringType.TypeName():
			ts, err := time.Parse(time.RFC3339, arg.Value().(string))
			if err != nil {
				return "", fmt.Errorf("invalid timestamp %s: %w", arg.Value().(string), err)
			}
			return fmt.Sprintf(`"%s"`, ts.In(cvt.timeLoc).Format(cvt.timeFormat)), nil
		case types.IntType.TypeName():
			return fmt.Sprintf(`FROM_UNIXTIME(%d)`, arg.Value().(int64)), nil
		default:
			return "", fmt.Errorf("unsupport %s in %s", arg.Type().TypeName(), name)
		}

	case overloads.StartsWith, overloads.EndsWith:
		op, _ := findReverse(name)
		arg0, err := cvt.convert(target)
		if err != nil {
			return "", err
		}
		arg1, err := cvt.convert(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s LIKE %s)", arg0, fmt.Sprintf(op, strings.Trim(arg1, `"`))), nil
	}

	return "", fmt.Errorf("unsupport func %s", name)
}

func findReverse(name string) (string, bool) {
	switch name {
	case operators.LogicalAnd:
		return "AND", true
	case operators.LogicalOr:
		return "OR", true
	case operators.LogicalNot:
		return "NOT", true
	case overloads.StartsWith:
		return `"%s%%"`, true
	case overloads.EndsWith:
		return `"%%%s"`, true
	}
	return operators.FindReverse(name)
}

func (cvt *rawConverter) Extend(opts ...Option) (*rawConverter, error) {
	ncvt := &rawConverter{}

	ncvt.sqlNameMap = make(map[string]string, len(cvt.sqlNameMap))
	for k, v := range cvt.sqlNameMap {
		ncvt.sqlNameMap[k] = v
	}
	ncvt.timeLoc = cvt.timeLoc
	ncvt.timeFormat = cvt.timeFormat
	ncvt.prefix = cvt.prefix

	ncvt.envOpts = make([]cel.EnvOption, len(cvt.envOpts))
	copy(ncvt.envOpts, cvt.envOpts)

	for _, opt := range opts {
		if err := opt(ncvt); err != nil {
			return nil, err
		}
	}

	env, err := NewEnv(append(ncvt.envOpts, cel.Variable(cvt.prefix, types.DynType))...)
	if err != nil {
		return nil, fmt.Errorf("new cel env: %w", err)
	}
	ncvt.Env = env

	return ncvt, nil
}
