package util

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

func SetJsonFields(json string, kvMap map[string]any) (string, error) {
	for key, val := range kvMap {
		newJson, err := sjson.Set(json, key, val)
		json = newJson
		if err != nil {
			return json, errors.WithStack(err)
		}
	}
	return json, nil
}

// 去除json中的转义字符
func DisableEscapeHtml(data interface{}) (string, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(data); err != nil {
		return "", err
	}
	return bf.String(), nil
}
