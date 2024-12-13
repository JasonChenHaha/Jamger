package jevent

import (
	"jconfig"
	"jglobal"
	"jlog"
	"log"

	"github.com/nsqio/go-nsq"
)

var ev *event

type LocalHandler func(context any)

type RemoteHandler func(message *nsq.Message) error

type event struct {
	localHandler map[uint32][]LocalHandler
	producer     *nsq.Producer
	consumer     map[string]*nsq.Consumer
}

// ------------------------- inside -------------------------

func init() {
	ev = &event{
		localHandler: map[uint32][]LocalHandler{},
		consumer:     map[string]*nsq.Consumer{},
	}
}

// ------------------------- outside -------------------------

// handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件
func LocalRegister(id uint32, handler LocalHandler) {
	ev.localHandler[id] = append(ev.localHandler[id], handler)
}

func LocalPublish(id uint32, context any) {
	if o, ok := ev.localHandler[id]; ok {
		for _, v := range o {
			v(context)
		}
	}
}

func RemoteRegister(id string, handler RemoteHandler) {
	consumer, err := nsq.NewConsumer(id, jglobal.SVR_NAME, nsq.NewConfig())
	if err != nil {
		jlog.Panic(err)
	}
	consumer.AddHandler(nsq.HandlerFunc(handler))
	consumer.SetLogger(jlog.Logger(), nsq.LogLevelWarning)
	err = consumer.ConnectToNSQLookupd(jconfig.GetString("nsq.lookupAddr"))
	if err != nil {
		jlog.Panic(err)
	}
	ev.consumer[id] = consumer
}

func RemotePublish(id string, data []byte) {
	if ev.producer == nil {
		producer, err := nsq.NewProducer(jconfig.GetString("nsq.addr"), nsq.NewConfig())
		if err != nil {
			jlog.Panic(err)
		}
		producer.SetLogger(jlog.Logger(), nsq.LogLevelWarning)
		ev.producer = producer
	}
	err := ev.producer.Publish(id, data)
	if err != nil {
		log.Panic(err)
	}
}
