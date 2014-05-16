package event

import "encoding/json"

type payload struct {
	Name string        `json:"name"`
	Args []interface{} `json:"args"`
}

func (p *payload) encode(dest *string) error {
	buf, err := json.Marshal(p)
	if err != nil {
		return err
	}
	*dest = string(buf)
	return nil
}

func (p *payload) decode(src string) error {
	if err := json.Unmarshal([]byte(src), &p); err != nil {
		return err
	}
	return nil
}
