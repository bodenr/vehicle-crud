package svr

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

type Encoding interface {
	Marshal(dataType interface{}) ([]byte, error)
	Unmarshal(data []byte, dataType interface{}) error
}

type JSONEncoding struct{}
type XMLEncoding struct{}

func (e *JSONEncoding) Marshal(dataType interface{}) ([]byte, error) {
	return json.Marshal(dataType)
}

func (e *JSONEncoding) Unmarshal(data []byte, dataType interface{}) error {
	return json.Unmarshal(data, dataType)
}

func (e *XMLEncoding) Marshal(dataType interface{}) ([]byte, error) {
	return xml.Marshal(dataType)
}

func (e *XMLEncoding) Unmarshal(data []byte, dataType interface{}) error {
	return xml.Unmarshal(data, dataType)
}

var encodings = map[string]Encoding{
	"application/json": &JSONEncoding{},
	"application/xml":  &XMLEncoding{},
}

func SupportsEncoding(encoding string) bool {
	_, exists := encodings[encoding]
	return exists
}

func Unmarshal(encoding string, data []byte, dataType interface{}) error {
	enc, exists := encodings[encoding]
	if !exists {
		return fmt.Errorf("No such encoding: %s", encoding)
	}
	return enc.Unmarshal(data, dataType)
}

func Marshal(encoding string, dataType interface{}) ([]byte, error) {
	enc, exists := encodings[encoding]
	if !exists {
		return nil, fmt.Errorf("No such encoding: %s", encoding)
	}
	return enc.Marshal(dataType)
}

func GetRequestContentType(request *http.Request) string {
	for _, contentType := range request.Header.Values("Content-Type") {
		if SupportsEncoding(contentType) {
			return contentType
		}
	}
	return ""
}

func GetResponseContentType(request *http.Request) string {
	// TODO: support weighted encodings, */*, etc..
	for _, acceptType := range request.Header.Values("Accept") {
		if SupportsEncoding(acceptType) {
			return acceptType
		}
	}
	// no Accept specified; if Content-Type was given use it
	contentType := GetRequestContentType(request)
	if contentType != "" {
		return contentType
	}
	// default to JSON
	return "application/json"
}
