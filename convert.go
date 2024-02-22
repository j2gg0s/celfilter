package celfilter

import (
	"fmt"

	"github.com/google/cel-go/cel"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
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

func convertLiteral(val ref.Val) string {
	switch val.Type().TypeName() {
	case types.StringType.TypeName():
		return fmt.Sprintf(`"%s"`, val.Value().(string))
	default:
		val = val.ConvertToType(types.StringType)
		return val.Value().(string)
	}
}

func convertCall(expr celast.CallExpr) (string, error) {
	name := expr.FunctionName()
	args := expr.Args()

	switch name {
	case operators.LogicalAnd, operators.LogicalOr,
		operators.Equals, operators.NotEquals,
		operators.Less, operators.LessEquals,
		operators.Greater, operators.GreaterEquals:

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

	case operators.LogicalNot:
		op, _ := findReverse(name)

		arg0, err := convert(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s)", op, arg0), nil
	}

	return "", fmt.Errorf("unsupport function: %s", name)
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
	}
	return operators.FindReverse(name)
}
