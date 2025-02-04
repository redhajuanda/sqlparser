# SQLParser

This repository is a clone of [github.com/vitessio/vitess](https://github.com/vitessio/vitess), specifically focusing on the SQL parser module. It provides an interface to parse SQL queries into structured statements for further processing.

## License
This project applies the same LICENSE as [vitessio/vitess](https://github.com/vitessio/vitess).

## Installation
```sh
go get github.com/redhajuanda/sqlparser
```

## Usage
```go
import (
	"github.com/redhajuanda/sqlparser"
)

func main() {
	ps, err := sqlparser.New(sqlparser.Options{})
	if err != nil {
		panic(err)
	}

	// Parse SQL query to sqlparser statement
	stmt, err := ps.Parse("SELECT * FROM employees")
	if err != nil {
		panic(err)
	}

	// Handle the statement based on the type
	switch stmt := stmt.(type) {
	case *sqlparser.Select:
		// Do something with SELECT statement
	}
}
```

## Comparison with Other Projects
There is an existing project, [xwb1989/sqlparser](https://github.com/xwb1989/sqlparser), which is also a standalone clone of the Vitess SQL parser. However, it is outdated and requires manual updates to stay in sync with the original Vitess repository. This project aims to provide an actively maintained and regularly updated version of the SQL parser

## Features
- Parses SQL queries into structured statements.
- Supports multiple SQL statement types.
- Provides an easy-to-use API for handling parsed statements.

## Acknowledgments
This project is based on the SQL parser module from [Vitess](https://vitess.io/).

## Important Notice
This repository will be regularly updated and synced with changes from the original repository. If there are any updates that I may have missed, please don't hesitate to comment or open an issue to bring it to my attention.
