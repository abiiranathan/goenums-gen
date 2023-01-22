# goenums

Generate go type-safe enumerated constants from sql code.

Works with postgresql CREATE TYPE typename AS ENUM statements
parsed from any .sql file to generate all constants, methods that
satisfy sql.Scanner interface method, driver.Valuer interface,
IsValid() method or a method for retrieving all valid values as a slice
of strings.

- Perfectly escapes both single-line and multiline comments.
- Generate constant using idiomatic go naming conventions.
- Simple implementation.
- Case insensitive to commands.

## Installation

```bash
go install github.com/abiiranathan/goenums@latest
```

### Usage:

Run goenums -help to see the usage.

```
go enums -in=types.sql -out=types.go -pkg=types
```

This will a file types.go with package name types in current directory by analysing an SQL file types.sql.

Example:

```sql
-- Tracks the status of a patient on the Theatre list.
CREATE TYPE operation_status AS ENUM(
  'PENDING',
  'ON GOING',
  'COMPLETED',
  'POSTPONED',
  'CANCELLED'
);

```

The generated code:

```go
package types


type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "PENDING"
	OperationStatusOnGoing   OperationStatus = "ON GOING"
	OperationStatusCompleted OperationStatus = "COMPLETED"
	OperationStatusPostponed OperationStatus = "POSTPONED"
	OperationStatusCancelled OperationStatus = "CANCELLED"
)

func (e OperationStatus) IsValid() bool {
	validValues := []string{
		"PENDING",
		"ON GOING",
		"COMPLETED",
		"POSTPONED",
		"CANCELLED",
	}

	for _, val := range validValues {
		if val == string(e) {
			return true
		}
	}
	return false
}

func (e OperationStatus) ValidValues() []string {
	return []string{
		"PENDING",
		"ON GOING",
		"COMPLETED",
		"POSTPONED",
		"CANCELLED",
	}
}

func (e *OperationStatus) Scan(src interface{}) error {
	source, ok := src.(string)
	if !ok {
		return fmt.Errorf("invalid value for %s: %s", "OperationStatus", source)
	}
	*e = OperationStatus(source)
	return nil
}

func (e OperationStatus) Value() (driver.Value, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid value for %s", "OperationStatus")
	}
	return string(e), nil
}


```

Happy Hacking!!
