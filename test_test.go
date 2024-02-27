package celfilter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Fixture struct {
	expr     string
	expected string
	opts     []Option
}

func newf(expr, expected string, opts ...Option) Fixture {
	return Fixture{
		expr:     expr,
		expected: expected,
		opts:     opts,
	}
}

func TestConvert(t *testing.T) {
	fixtures := []Fixture{
		newf(
			`i.name == "j2gg0s"`,
			`(name == "j2gg0s")`,
		),
		newf(
			`i.name != "j2gg0s"`,
			`(name != "j2gg0s")`,
		),
		newf(
			`i.age < 18`,
			`(age < 18)`,
		),
		newf(
			`i.age + 10 < 28 `,
			`((age + 10) < 28)`,
		),
		newf(
			`i.name == "j2gg0s" && i.age > 18`,
			`((name == "j2gg0s") AND (age > 18))`,
		),
		newf(
			`i.name == "j2gg0s" && (i.age > 18 || i.city == "上海")`,
			`((name == "j2gg0s") AND ((age > 18) OR (city == "上海")))`,
		),
		newf(
			`i.name == "j2gg0s" && i.age > 18 && i.city == "上海"`,
			`(((name == "j2gg0s") AND (age > 18)) AND (city == "上海"))`,
		),
		newf(
			`i.name.startsWith("j")`,
			`(name LIKE "j%")`,
		),
		newf(
			`i.name.endsWith("0s")`,
			`(name LIKE "%0s")`,
		),
		newf(
			`i.birthtime < timestamp("2024-01-01T00:00:00Z")`, // RFC3339
			`(birthtime < "2024-01-01 08:00:00")`,             // use system Local timezone, current is +08
			WithLocation(time.FixedZone("Asia/Shanghai", 8*int(time.Hour/time.Second))),
		),
		newf(
			`i.birthtime < timestamp(1704067200)`,     // UTC seconds
			`(birthtime < FROM_UNIXTIME(1704067200))`, // use from_unixstamp
		),
		newf(
			`-i.age > -18`,
			`((- age) > -18)`,
		),
		newf(
			`!i.joined`,
			`(NOT joined)`,
		),
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.expr, func(t *testing.T) {
			sql, err := Convert(fixture.expr, fixture.opts...)
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, sql) //nolint:staticcheck
		})
	}
}
