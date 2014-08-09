package bind

import "fmt"

type Error struct {
	Field   string
	Message string
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Field, err.Message)
}
