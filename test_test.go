package celfilter

import (
	"testing"
	"time"

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
			`i.age + 10 < 28 `,
			`((age + 10) < 28)`,
		},
		{
			`i.name == "j2gg0s" && i.age > 18`,
			`((name == "j2gg0s") AND (age > 18))`,
		},
		{
			`i.name == "j2gg0s" && (i.age > 18 || i.city == "上海")`,
			`((name == "j2gg0s") AND ((age > 18) OR (city == "上海")))`,
		},
		{
			`i.name == "j2gg0s" && i.age > 18 && i.city == "上海"`,
			`(((name == "j2gg0s") AND (age > 18)) AND (city == "上海"))`,
		},
		{
			`i.name.startsWith("j")`,
			`(name LIKE "j%")`,
		},
		{
			`i.name.endsWith("0s")`,
			`(name LIKE "%0s")`,
		},
		{
			`i.birthtime < timestamp("2024-01-01T00:00:00Z")`, // RFC3339
			`(birthtime < "2024-01-01 08:00:00")`,             // use system Local timezone, current is +08
		},
		{
			`i.birthtime < timestamp(1704067200)`,     // UTC seconds
			`(birthtime < FROM_UNIXTIME(1704067200))`, // use from_unixstamp
		},
		{
			`-i.age > -18`,
			`((- age) > -18)`,
		},
		{
			`!i.joined`,
			`(NOT joined)`,
		},
	}

	time.Local = time.FixedZone("Asia/Shanghai", int(8*time.Hour/time.Second))
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.expr, func(t *testing.T) {
			t.Parallel()
			ast, iss := env.Compile(fixture.expr)
			assert.Nil(t, iss)

			sql, err := Convert(ast)
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, sql, ast.Expr().String()) //nolint:staticcheck
		})
	}
}
