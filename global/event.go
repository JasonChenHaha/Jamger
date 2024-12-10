package jglobal

type Handler func(context any)

var Event *event

type event struct {
	handler map[uint32][]Handler
}

// ------------------------- inside -------------------------

func init() {
	Event = &event{handler: map[uint32][]Handler{}}
}

// handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件
func (ev *event) Register(id uint32, handler Handler) {
	ev.handler[id] = append(ev.handler[id], handler)
}

func (ev *event) Emit(id uint32, context any) {
	if o, ok := ev.handler[id]; ok {
		for _, v := range o {
			v(context)
		}
	}
}
