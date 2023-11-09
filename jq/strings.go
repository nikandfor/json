package jq

type (
	Cat struct {
		Filters []Filter
	}
)

func (f Cat) Next(w, r []byte, st int, state *State) ([]byte, int, *State, error) {
	return w, st, nil, nil
}
