WIPS

Using a subset of [CEL](https://github.com/google/cel-go) to express filter conditions.
And support translate them to MySQL/PostgreSQL.

We currently only support the following operators.
```
&&    LogicalAnd
||    LogicalOr
!     LogicalNot
==    Equals
!=    NotEquals
<     Less
<=    LessEquals
>     Greater
>=    GreaterEquals
+     Add
-     Subtract
*     Multiply
/     Divide
%     Modulo
-     Negate
startsWith        MemberFunction of String
endsWith          MemberFunction of String
timestamp         Convert string/int to Timestamp
```
