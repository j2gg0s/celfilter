# celfilter
使用 [CEL](https://github.com/google/cel-spec/blob/master/doc/intro.md) 的子集表达过滤条件, 并支持转化到对应的数据库查询条件.

当前支持的语言子集包括:
- 基础的数据类型, 包括 `int`, `double`, `bool`, `string`, `bytes`
- 算术运算符, `+`, `-`, `*`, `/`, `%`
- 逻辑运算符, `&&`, `||`, `!`
- 比较运算符, `==`, `!=`, `>`, `>=`, `<`, `<=`
- 使用 `()` 控制优先级
- startsWith 和 endsWith
- 时间转换函数 timestamp

转换案例:
```go
package celfilter

import "fmt"

func ExampleConvert() {
	mustPrint := func(s string, err error) {
		if err != nil {
			panic(nil)
		}
		fmt.Println(s)
	}

	// i 为默认前缀
	mustPrint(Convert(`i.name == "j2gg0s"`))
	mustPrint(Convert(`i.age > 25`))
	mustPrint(Convert(`i.name == "j2gg0s" && i.age > 25`))
	mustPrint(Convert(`i.name == "j2gg0s" && (i.city == "Shanghai" || i.city == "SuZhou")`))
	mustPrint(Convert(`i.name.startsWith("j2")`))
	mustPrint(Convert(`i.name.endsWith("0s")`))
	mustPrint(Convert(`i.birthday == timestamp("2000-01-01T00:00:00.000Z")`))
	mustPrint(Convert(`i.birthday == timestamp(1704067200)`))

	// Output: (name == "j2gg0s")
	// (age > 25)
	// ((name == "j2gg0s") AND (age > 25))
	// ((name == "j2gg0s") AND ((city == "Shanghai") OR (city == "SuZhou")))
	// (name LIKE "j2%")
	// (name LIKE "%0s")
	// (birthday == "2000-01-01 08:00:00")
	// (birthday == FROM_UNIXTIME(1704067200))
}
```

## 限制及注意点
### timestamp
仅接受 string 和 int 做为参数.

当参数为 string 时, 视作格式为 YYYY-MM-DDThh:mm:ssTZD 的时间字符串, 即 Go 中的 [time.RFC3339](https://pkg.go.dev/time#pkg-constants).
其中:
- YYYY  表示年份, 4 位数字
- MM    表示月份, 2 位数字
- DD    表示日期, 2 位数字
- T     是日期和时间的分隔符
- hh    表示小时, 2 位数字
- mm    表示分钟, 2 位数字
- ss    表示秒, 2 位数字
- TZD   表示时区偏移, 可以是 "Z" 或者是 "[+/-]hh"

此时, celfilter 会将其解析为时间后转换成用户指定的时区, 默认为系统时区.
请结合数据库时区等因素指定正确时区.

当参数位 int 时, 视作 UNIX 时间戳. 即自 1970-01-01T00:00:00Z 起经过的秒数.
此时, celfilter 会将其转换为对数据 FROM_UNIXTIME 函数的调用.

你也可以直接使用字符串而不实用 timestamp 做转换, 此时需要使用数据库对应的字符串格式.
