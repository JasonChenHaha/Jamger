package jevent

import (
	"jconfig"
	"jglobal"
	"jlog"
	"log"

	"github.com/nsqio/go-nsq"
)

const (
	EVENT_TEST_1 = 0
	EVENT_TEST_2 = "jamger"
)

type LocalHandler func(context any) // handler有义务将高耗时逻辑放入协程中处理，防止delay后续事件

type RemoteHandler func(message *nsq.Message) error

type event struct {
	localHandler map[uint32][]LocalHandler
	producer     *nsq.Producer
	consumer     map[string]*nsq.Consumer
}

var Event *event

// ------------------------- outside -------------------------

func Init() {
	Event = &event{
		localHandler: map[uint32][]LocalHandler{},
		consumer:     map[string]*nsq.Consumer{},
	}
}

func (o *event) LocalRegister(id uint32, handler LocalHandler) {
	o.localHandler[id] = append(o.localHandler[id], handler)
}

func (o *event) LocalPublish(id uint32, context any) {
	if v, ok := o.localHandler[id]; ok {
		for _, v := range v {
			v(context)
		}
	}
}

func (o *event) RemoteRegister(id string, handler RemoteHandler) {
	consumer, err := nsq.NewConsumer(id, jglobal.SERVER, nsq.NewConfig())
	if err != nil {
		jlog.Panic(err)
	}
	consumer.AddHandler(nsq.HandlerFunc(handler))
	consumer.SetLogger(jlog.Logger(), nsq.LogLevelWarning)
	err = consumer.ConnectToNSQLookupd(jconfig.GetString("nsq.lookupAddr"))
	if err != nil {
		jlog.Panic(err)
	}
	o.consumer[id] = consumer
}

func (o *event) RemotePublish(id string, data []byte) {
	if o.producer == nil {
		producer, err := nsq.NewProducer(jconfig.GetString("nsq.addr"), nsq.NewConfig())
		if err != nil {
			jlog.Panic(err)
		}
		producer.SetLogger(jlog.Logger(), nsq.LogLevelWarning)
		o.producer = producer
	}
	err := o.producer.Publish(id, data)
	if err != nil {
		log.Panic(err)
	}
}
