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
	id     uint8
}

type Response interface {
	GetStatus() Status
}

func (r response[T]) GetStatus() Status { return r.status }

func (r response[T]) String() string {
	var body string
	switch r.status {
	case Ok:
		body = fmt.Sprintf("%s,", r.status)
	case NotFound, ServerError:
		body = r.status.String()
	default:
		panic("Unreachable")
	}
	return fmt.Sprintf("Response(%d)<%s>", r.id, body)
}

func ConstructResponse[T types.IntOrString](request Request, status Status, data T) Response {
	switch v := any(data).(type) {
	case int:
		return response[int]{status: status, data: v, id: request.GetId()}
	case string:
		return response[string]{status: status, data: v, id: request.GetId()}
	}
	panic("Unreachable")
}
