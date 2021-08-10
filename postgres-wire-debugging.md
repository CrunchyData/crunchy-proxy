
# Parsing a raw column message in postgres wire protocol

Reference:  https://www.postgresql.org/docs/current/protocol-message-formats.html

## T column message

```
84 0 0 0 77 0 3 105 100 0 0 0 67 127 0 1 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

84 is 'T'  (byte)

0 0 0 77 is the length (77)  (Int32)

That leaves us with :
0 3 105 100 0 0 0 67 127 0 1 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

0 3 is the number of fields in a row (3) (Int16)

That leaves us with :

105 100 0 0 0 67 127 0 1 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Then, for each of those 3 columns...

105 100 0 is 'id' and then null.  (String)

That leaves us with :

0 0 67 127 0 1 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int32
If the field can be identified as a column of a specific table, the object ID of the table; otherwise zero.

data = 0 0 67 127
left = 0 1 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int16
If the field can be identified as a column of a specific table, the attribute number of the column; otherwise zero.

data = 0 1, so the first column
left = 0 0 0 23 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int32
The object ID of the field's data type.

data = 0 0 0 23, so that must be bytea in postgres
left = 0 4 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int16
The data type size (see pg_type.typlen). Note that negative values denote variable-width types.

data = 0 4
left = 255 255 255 255 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

data = 255 255 255 255
left = 0 0 117 115 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int32
The type modifier (see pg_attribute.atttypmod). The meaning of the modifier is type-specific.

data = 0 0 117 115
left = 101 114 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

Int16
The format code being used for the field. Currently will be zero (text) or one (binary). In a RowDescription returned from the statement variant of Describe, the format code is not yet known and will always be zero.

data = 101 114
left = 110 97 109 101 0 0 0 67 127 0 2 0 0 4 19 255 255 0 0 0 84 0 0 100 97 116 97 0 0 0 67 127 0 3 0 0 0 17 255 255 255 255 255 255 0

**next column!**

data = 110 97 109 101 0, or "name", plus a null

and process repeats.
```

## Data Message

```
68 0 0 0 45 0 3 0 0 0 1 49 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

68 is 'D' (byte)
left = 0 0 0 45 0 3 0 0 0 1 49 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

0 0 0 45 is the length (45) (Int32)
left = 0 3 0 0 0 1 49 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

0 3 is the number of columns (3) (Int16)
left = 0 0 0 1 49 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

Then, for each column...

0 0 0 1 is the length of the column value (1) (Int32)
left = 49 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

NOTE, N.B.!!  Can be zero. As a special case, -1 indicates a NULL column value. No value bytes follow in the NULL case.

49 is the value of the column, in this case '1' (id of '1') (byte * n from the previous step)

left = 0 0 0 6 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

length is 0 0 0 6 (Int32)
left = 102 111 111 98 97 114 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

102 111 111 98 97 114 is 'foobar'
left == 0 0 0 20 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

length is 0 0 0 20 (Int32)
left = 92 120 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49 52 49

This is:

['\\', 'x', '4', '1', '4', '1', '4', '1', '4', '1', '4', '1', '4', '1', '4', '1', '4', '1', '4', '1']

So an escaped binary sorta syntax.
```



