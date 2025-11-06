package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"

	"github.com/EnemigoPython/go-getit/src/types"
)

type Status byte

const (
	Ok Status = iota
	NotFound
	StreamDone
	ServerError
)

func (s Status) String() string {
	return [...]string{"Ok", "NotFound", "StreamDone", "ServerError"}[s]
}

func (s Status) ToLower() string {
	return [...]string{"ok", "notfound", "streamdone", "servererror"}[s]
}

type response[T types.IntOrString] struct {
	status   Status
	data     T
	hasData  bool
	id       uint8
	isStream bool
}

type Response interface {
	GetStatus() Status
	StreamDone() bool
	EncodeResponse() []byte
	DataPayload() string
}

func (r response[T]) GetStatus() Status { return r.status }

func (r response[T]) StreamDone() bool {
	return r.isStream && r.status == Ok
}

func (r response[T]) String() string {
	var body string
	if r.hasData {
		switch d := any(r.data).(type) {
		case int:
			body = fmt.Sprintf("%s,%d", r.status, d)
		case string:
			body = fmt.Sprintf("%s,%s", r.status, d)
		}
	} else {
		body = r.status.String()
	}
	return fmt.Sprintf("Response(%d)<%s>", r.id, body)
}

func (r response[T]) EncodeResponse() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(r.status))
	if r.status != Ok || !r.hasData {
		return buf.Bytes()
	}
	switch d := any(r.data).(type) {
	case int:
		buf.WriteByte(byte(0)) // type of data: int
		binary.Write(buf, binary.BigEndian, int32(d))
	case string:
		buf.WriteByte(byte(1)) // type of data: string
		buf.Write([]byte(d))
	}
	return buf.Bytes()
}

func (r response[T]) DataPayload() string {
	if r.status == ServerError {
		log.Fatal("Server error")
	}
	if r.status == NotFound {
		return "" // impossible value
	}
	switch d := any(r.data).(type) {
	case int:
		return strconv.Itoa(d)
	case string:
		return d
	}
	panic("Unreachable")
}

func ConstructResponse[T types.IntOrString](request Request, status Status, data T) Response {
	isStream := request.IsStream()
	var hasData bool
	switch request.GetAction() {
	case Store, Load, Keys, Count:
		hasData = true
	}
	switch v := any(data).(type) {
	case int:
		return response[int]{
			status:   status,
			data:     v,
			id:       request.GetId(),
			hasData:  hasData,
			isStream: isStream,
		}
	case string:
		return response[string]{
			status:   status,
			data:     v,
			id:       request.GetId(),
			hasData:  hasData,
			isStream: isStream,
		}
	}
	panic("Unreachable")
}

func DecodeResponse(b []byte) Response {
	status := Status(b[0])
	if status != Ok || len(b) < 2 {
		return response[int]{status: status, hasData: false}
	}
	if b[1] == 0 {
		data := int32(binary.BigEndian.Uint32(b[2:]))
		return response[int]{
			status:  status,
			data:    int(data),
			hasData: true,
		}
	} else {
		data := string(b[2:])
		return response[string]{
			status:  status,
			data:    data,
			hasData: true,
		}
	}
}
