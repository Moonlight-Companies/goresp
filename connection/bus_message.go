package connection

import "encoding/json"

type BusMessage struct {
	Channel string
	Data    []byte
	Pattern string
}

func (m *BusMessage) IntoMap() (output map[string]interface{}, err error) {
	if err = json.Unmarshal(m.Data, &output); err != nil {
		return nil, err
	}
	return output, nil
}
