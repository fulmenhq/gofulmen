package fulencode

import "fmt"

type Operation string

const (
	OperationEncode    Operation = "encode"
	OperationDecode    Operation = "decode"
	OperationDetect    Operation = "detect"
	OperationNormalize Operation = "normalize"
	OperationBOM       Operation = "bom"
)

// FulencodeError implements the canonical fulencode error envelope.
//
// Schema: schemas/crucible-go/library/fulencode/v1.0.0/fulencode-error.schema.json
type FulencodeError struct {
	Code         string          `json:"code"`
	Message      string          `json:"message"`
	Operation    Operation       `json:"operation"`
	InputFormat  *EncodingFormat `json:"input_format,omitempty"`
	OutputFormat *EncodingFormat `json:"output_format,omitempty"`
	Details      map[string]any  `json:"details,omitempty"`
}

func (e *FulencodeError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func newError(op Operation, code, msg string) *FulencodeError {
	return &FulencodeError{Code: code, Message: msg, Operation: op}
}

func withInputFormat(err *FulencodeError, format EncodingFormat) *FulencodeError {
	if err != nil {
		err.InputFormat = &format
	}
	return err
}

func withOutputFormat(err *FulencodeError, format EncodingFormat) *FulencodeError {
	if err != nil {
		err.OutputFormat = &format
	}
	return err
}

func addDetails(err *FulencodeError, kv map[string]any) *FulencodeError {
	if err == nil || len(kv) == 0 {
		return err
	}
	if err.Details == nil {
		err.Details = make(map[string]any, len(kv))
	}
	for k, v := range kv {
		err.Details[k] = v
	}
	return err
}
