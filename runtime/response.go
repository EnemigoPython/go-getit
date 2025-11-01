package runtime

import (
	"fmt"

	"github.com/EnemigoPython/go-getit/types"
)

type Status byte

const (
	Ok Status = iota
	NotFound
	ServerError
)

func (s Status) String() string {
	return [...]string{"Ok", "NotFound", "ServerError"}[s]
}

func (s Status) ToLower() string {
	return [...]string{"ok", "notfound", "servererror"}[s]
}

type response[T types.IntOrString] struct {
	status Status
	data   T
}

type Response interface {
	GetResponseStatus() Status
}

func (r response[T]) GetResponseStatus() Status {
	return r.status
}

func (r response[T]) String() string {
	var body string
	switch r.status {
	case Ok:
		body = fmt.Sprintf("%s,", r.status)
	case NotFound:
		body = fmt.Sprintf("%s,", r.status)
	case ServerError:
		body = fmt.Sprintf("%s,", r.status)
	default:
		panic("Unreachable")
	}
	return fmt.Sprintf("Response<%s>", body)
}

func ConstructResponse(status Status) Response {
	return response[int]{data: 0}
}
