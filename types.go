package birdactyl

import "encoding/json"

type Event struct {
	Type string
	Data map[string]string
	Sync bool
}

type EventResult struct {
	allow   bool
	message string
}

func Allow() EventResult {
	return EventResult{allow: true}
}

func Block(message string) EventResult {
	return EventResult{allow: false, message: message}
}

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Query   map[string]string
	Body    map[string]interface{}
	RawBody []byte
	UserID  string
}

type Response struct {
	Status  int
	Headers map[string]string
	body    []byte
}

func JSON(data interface{}) Response {
	b, _ := json.Marshal(map[string]interface{}{"success": true, "data": data})
	return Response{Status: 200, Headers: map[string]string{"Content-Type": "application/json"}, body: b}
}

func Error(status int, msg string) Response {
	b, _ := json.Marshal(map[string]interface{}{"success": false, "error": msg})
	return Response{Status: status, Headers: map[string]string{"Content-Type": "application/json"}, body: b}
}

func Text(text string) Response {
	return Response{Status: 200, Headers: map[string]string{"Content-Type": "text/plain"}, body: []byte(text)}
}

func (r Response) WithStatus(status int) Response {
	r.Status = status
	return r
}

func (r Response) WithHeader(key, value string) Response {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
	return r
}
