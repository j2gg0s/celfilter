package celfilter

import (
	"fmt"
	"time"

	"github.com/google/cel-go/cel"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/overloads"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func Convert(ast *cel.Ast) (string, error) {
	return convert(ast.NativeRep().Expr())
}

func convert(expr celast.Expr) (string, error) {
	switch expr.Kind() {
	case celast.CallKind:
		return convertCall(expr.AsCall())
	case celast.SelectKind:
		return convertSelect(expr.AsSelect())
	case celast.LiteralKind:
		return convertLiteral(expr.AsLiteral()), nil
	}
	return "", nil
}

func convertCall(expr celast.CallExpr) (string, error) {
	name := expr.FunctionName()
	target := expr.Target()
	args := expr.Args()

	switch name {
	case operators.LogicalAnd, operators.LogicalOr,
		operators.Equals, operators.NotEquals,
		operators.Less, operators.LessEquals,
		operators.Greater, operators.GreaterEquals,
		operators.Add, operators.Subtract,
		operators.Multiply, operators.Divide,
		operators.Modulo:

		op, _ := findReverse(name)
		arg0, err := convert(args[0])
		if err != nil {
			return "", err
		}
		arg1, err := convert(args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", arg0, op, arg1), nil

	case operators.LogicalNot,
		operators.Negate:

		op, _ := findReverse(name)

		arg0, err := convert(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s)", op, arg0), nil

	case overloads.StartsWith, overloads.EndsWith:
		op, _ := findReverse(name)
		ins, err := convert(target)
		if err != nil {
			return "", err
		}
		if args[0].Kind() != celast.LiteralKind {
			return "", fmt.Errorf("except literal in %s: %v", op, args[0].Kind())
		}
		arg := args[0].AsLiteral()
		if arg.Type().TypeName() != types.StringType.TypeName() {
			return "", fmt.Errorf("except string: %s", arg.Type().TypeName())
		}
		return fmt.Sprintf(
			"(%s LIKE %s)", ins, fmt.Sprintf(op, arg.Value().(string))), nil

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
			return fmt.Sprintf(`"%s"`, ts.Local().Format("2006-01-02 15:04:05")), nil
		case types.IntType.TypeName():
			return fmt.Sprintf(`FROM_UNIXTIME(%d)`, arg.Value().(int64)), nil
		default:
			return "", fmt.Errorf("unsupport %s in %s", arg.Type().TypeName(), name)
		}

	}

	return "", fmt.Errorf("unsupport function: %s", name)
}

func convertLiteral(val ref.Val) string {
	switch val.Type().TypeName() {
	case types.StringType.TypeName():
		return fmt.Sprintf(`"%s"`, val.Value().(string))
	default:
		val = val.ConvertToType(types.StringType)
		return val.Value().(string)
	}
}

func convertSelect(expr celast.SelectExpr) (string, error) {
	op := expr.Operand()
	if op.Kind() != celast.IdentKind && op.AsIdent() != "i" {
		return "", fmt.Errorf("unsupport ident: %v", expr)
	}
	return expr.FieldName(), nil
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
