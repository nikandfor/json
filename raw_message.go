package json

type RawMessage []byte

func (x RawMessage) MarshalJSON() ([]byte, error) {
	return x, nil
}

func (x *RawMessage) UnmarshalJSON(d []byte) error {
	*x = append((*x)[:0], d...)

	return nil
}
