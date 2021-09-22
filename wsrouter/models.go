package wsrouter

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/zanloy/bms-api/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// "borrowed" from https://stackoverflow.com/a/36418082
type Consumer struct {
	Connection net.Conn
	Filter     *models.Filter
	Source     chan interface{}
	Quit       chan struct{}
}

func (consumer *Consumer) Start() {
	consumer.Source = make(chan interface{}, 10)
	go func() {
		for {
			select {
			case msg := <-consumer.Source:
				if consumer.Filter != nil {
					if obj, ok := msg.(unstructured.Unstructured); ok {
						if consumer.Filter.Kind != "" && obj.GetKind() != consumer.Filter.Kind {
							continue
						}
						if consumer.Filter.Namespace != "" && obj.GetNamespace() != consumer.Filter.Namespace {
							continue
						}
						if consumer.Filter.Name != "" && obj.GetName() != consumer.Filter.Name {
							continue
						}
					}
					consumer.send(msg)
				} else {
					consumer.send(msg)
				}
			case <-consumer.Quit:
				return
			}
		}
	}()
}

func (consumer *Consumer) send(msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		logger.Err(err).Msg("error while trying to marshal msg object")
		return
	}
	err = wsutil.WriteServerMessage(consumer.Connection, ws.OpText, bytes)
	if err != nil {
		logger.Err(err).Msg("error while trying to send msg through websocket")
		return
	}
}

type channel struct {
	sync.Mutex
	consumers []*Consumer
}

func (ch *channel) Iter(routine func(*Consumer)) {
	ch.Lock()
	defer ch.Unlock()

	for _, consumer := range ch.consumers {
		routine(consumer)
	}
}

func (ec *channel) Push(consumer *Consumer) {
	ec.Lock()
	defer ec.Unlock()

	ec.consumers = append(ec.consumers, consumer)
}
