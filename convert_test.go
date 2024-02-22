package celfilter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	env, err := NewEnv()
	assert.NoError(t, err)

	fixtures := []struct {
		expr     string
		expected string
	}{
		{
			`i.name == "j2gg0s"`,
			`(name == "j2gg0s")`,
		},
		{
			`i.age < 18`,
			`(age < 18)`,
		},
		{
			`i.name == "j2gg0s" && i.age > 18`,
			`((name == "j2gg0s") AND (age > 18))`,
		},
		{
			`!i.joined`,
			`(NOT joined)`,
		},
		{
			`i.name == "j2gg0s" && (i.age > 18 || i.city == "上海")`,
			`((name == "j2gg0s") AND ((age > 18) OR (city == "上海")))`,
		},
		{
			`i.name == "j2gg0s" && i.age > 18 && i.city == "上海"`,
			`(((name == "j2gg0s") AND (age > 18)) AND (city == "上海"))`,
		},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.expr, func(t *testing.T) {
			t.Parallel()
			ast, iss := env.Compile(fixture.expr)
			assert.NotNil(t, iss)
			assert.NoError(t, iss.Err())

			sql, err := Convert(ast)
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, sql, ast.Expr().String()) //nolint:staticcheck
		})
	}
}
