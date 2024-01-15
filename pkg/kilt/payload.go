package kilt

import (
	"fmt"

	"github.com/go-akka/configuration/hocon"
)

func retrievePayload(object *hocon.HoconObject) (*Payload, error) {
	payload := new(Payload)
	payload.Type = Unknown
	if object.GetKey("url").IsString() {
		payload.Contents = object.GetKey("url").GetString()
		payload.Type = URL
	} else if object.GetKey("file").IsString() {
		payload.Contents = object.GetKey("file").GetString()
		payload.Type = LocalPath
	} else if object.GetKey("payload").IsString() {
		payload.Contents = object.GetKey("payload").GetString()
		payload.Type = Base64
	} else if object.GetKey("text").IsString() {
		payload.Contents = object.GetKey("text").GetString()
		payload.Type = Text
	}
	if object.GetKey("gzipped") != nil && object.GetKey("gzipped").IsString() && object.GetKey("gzipped").GetBoolean() {
		payload.Gzipped = true
	}

	if payload.Type == Unknown {
		return nil, fmt.Errorf("could not identify payload type for %s", object.ToString(1))
	}

	return payload, nil
}
