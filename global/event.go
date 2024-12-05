package jglobal

var Event *event

type event struct {
}

// ------------------------- inside -------------------------

func init() {
	Event = &event{}
}
