package runtime

import "github.com/EnemigoPython/go-getit/types"

type ResponseStatus byte

const (
	Ok ResponseStatus = iota
	NotFound
	ServerError
)

type response[T types.IntOrString] struct {
	action Action
	key    string
	data   T
}
