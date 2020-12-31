package svr

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

const (
	// ContentAppJSON http content type for json.
	ContentAppJSON = "application/json"

	// ContentAppXML http content type for xml.
	ContentAppXML = "application/xml"

	// ContentAppProtobuf content type for protobuf.
	ContentAppProtobuf = "application/x-protobuf"
)

// Encoding provides the means to marshal and unmarshal data.
type Encoding interface {
	Marshal(dataType interface{}) ([]byte, error)
	Unmarshal(data []byte, dataType interface{}) error
}

// JSONEncoding provides Encoding for JSON content.
type JSONEncoding struct{}

// XMLEncoding provides Encoding for XML content.
type XMLEncoding struct{}

// ProtobufEncoding provides Encoding for Protobuf content.
type ProtobufEncoding struct{}

// Marshal a data object into JSON.
func (e *JSONEncoding) Marshal(dataType interface{}) ([]byte, error) {
	return json.Marshal(dataType)
}

// Unmarshal JSON content into a data object.
func (e *JSONEncoding) Unmarshal(data []byte, dataType interface{}) error {
	return json.Unmarshal(data, dataType)
}

// Marshal a data object into XML.
func (e *XMLEncoding) Marshal(dataType interface{}) ([]byte, error) {
	return xml.Marshal(dataType)
}

// Unmarshal XML content into a data object.
func (e *XMLEncoding) Unmarshal(data []byte, dataType interface{}) error {
	return xml.Unmarshal(data, dataType)
}

// Marshal is not implemented an only provided to signify protobuf content is supported.
func (e *ProtobufEncoding) Marshal(dataType interface{}) ([]byte, error) {
	return nil, fmt.Errorf("No generic protobuf marshaller")
}

// Unmarshal is not implemented an only provided to signify protobuf content is supported.
func (e *ProtobufEncoding) Unmarshal(data []byte, dataType interface{}) error {
	return fmt.Errorf("No generic protobuf unmarshaller")
}

var encodings = map[string]Encoding{
	ContentAppJSON: &JSONEncoding{},
	ContentAppXML:  &XMLEncoding{},
	// TODO: find a way to support a generic protobuf encoding
	ContentAppProtobuf: &ProtobufEncoding{},
}

// SupportsEncoding returns if the said encoding type is supported.
func SupportsEncoding(encoding string) bool {
	_, exists := encodings[encoding]
	return exists
}

// Unmarshal the said encoding.
func Unmarshal(encoding string, data []byte, dataType interface{}) error {
	enc, exists := encodings[encoding]
	if !exists {
		return fmt.Errorf("No such encoding: %s", encoding)
	}
	return enc.Unmarshal(data, dataType)
}

// Marshal the said encoding.
func Marshal(encoding string, dataType interface{}) ([]byte, error) {
	enc, exists := encodings[encoding]
	if !exists {
		return nil, fmt.Errorf("No such encoding: %s", encoding)
	}
	return enc.Marshal(dataType)
}

// GetRequestContentType get the content type from the request header by checking it against supported encodings.
func GetRequestContentType(request *http.Request) string {
	for _, contentType := range request.Header.Values("Content-Type") {
		if SupportsEncoding(contentType) {
			return contentType
		}
	}
	return ""
}

// GetResponseContentType get the content type from the request header by checking it against supported encodings.
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
	return ContentAppJSON
}
