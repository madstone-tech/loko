package usecases

import (
	"example.com/fixture/cmd"
)

// BadImport calls cmd which is forbidden for usecases.
func BadImport() string {
	return cmd.Hello()
}
