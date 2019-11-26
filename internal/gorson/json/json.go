package json

import (
	"bytes"
	"encoding/json"
	"log"
)

func Marshal(parameters map[string]string) string {
	// we use a custom encoder here because the standard library
	// json.Marshal cannot be configured not to escape characters like
	// & < >
	// https://stackoverflow.com/a/28596225
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(&parameters); err != nil {
		log.Fatal(err)
	}
	ibuf := new(bytes.Buffer)
	err := json.Indent(ibuf, buf.Bytes(), "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	return ibuf.String()
}
