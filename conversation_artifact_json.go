package agentsdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func decodeArtifactEdit(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("artifact edit must be a single JSON object")
	}
	return nil
}

// Empty replacement and null cell value are explicit deletion requests.
// Missing properties must never be interpreted as those zero values.
func (e *ConversationArtifactTextEdit) UnmarshalJSON(raw []byte) error {
	var value struct {
		Find    *string `json:"find"`
		Replace *string `json:"replace"`
	}
	if err := decodeArtifactEdit(raw, &value); err != nil {
		return err
	}
	if value.Find == nil || value.Replace == nil {
		return errors.New("artifact text edit requires find and replace")
	}
	e.Find, e.Replace = *value.Find, *value.Replace
	return nil
}

func (e *ConversationArtifactCellEdit) UnmarshalJSON(raw []byte) error {
	var value struct {
		Row    *int            `json:"row"`
		Column *string         `json:"column"`
		Value  json.RawMessage `json:"value"`
	}
	if err := decodeArtifactEdit(raw, &value); err != nil {
		return err
	}
	if value.Row == nil || value.Column == nil || len(value.Value) == 0 {
		return errors.New("artifact cell edit requires row, column and value")
	}
	var cell *string
	if err := json.Unmarshal(value.Value, &cell); err != nil {
		return err
	}
	e.Row, e.Column, e.Value = *value.Row, *value.Column, cell
	return nil
}
