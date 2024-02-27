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
