package sp

import (
	"sync"
)

//Broker Broker
type Broker struct {
	slots []*slot
}

//NewBroker NewBroker
func NewBroker(slotCount uint32, broadcast func(topicID string, to map[string]*Consumer, from Session, message interface{})) *Broker {
	broker := Broker{}
	for i := uint32(0); i < slotCount; i++ {
		s := newSlot(broadcast)
		broker.slots = append(broker.slots, s)
	}
	return &broker
}

//Subscribe Subscribe
func (broker *Broker) Subscribe(topicID string, consumer *Consumer) {
	s := broker.slots[SlotBalance(topicID, uint32(len(broker.slots)))]
	s.subscribe(topicID, consumer)
}

//Publish Publish
func (broker *Broker) Publish(topicID string, from Session, message interface{}) {
	s := broker.slots[SlotBalance(topicID, uint32(len(broker.slots)))]
	s.publish(topicID, from, message)
}

//Unsubscribe Unsubscribe
func (broker *Broker) Unsubscribe(topicID string, consumerID string) {
	s := broker.slots[SlotBalance(topicID, uint32(len(broker.slots)))]
	s.unsubscribe(topicID, consumerID)
}

//GetTotalTopic GetTotalTopic
func (broker *Broker) GetTotalTopic() uint64 {
	count := uint64(0)
	for _, s := range broker.slots {
		count += s.topicCount()
	}
	return count
}

//GetTotalConsumer GetTotalConsumer
func (broker *Broker) GetTotalConsumer() map[string]uint64 {
	count := map[string]uint64{}
	for _, s := range broker.slots {
		c := s.consumerCount()
		for k, v := range c {
			count[k] = v
		}
	}
	return count
}

type slot struct {
	broadcast func(topicID string, to map[string]*Consumer, from Session, message interface{})
	topicMut  sync.RWMutex
	slotMut   sync.RWMutex
	topics    map[string]*topic
}

func newSlot(broadcast func(topicID string, to map[string]*Consumer, from Session, message interface{})) *slot {
	return &slot{topics: make(map[string]*topic), broadcast: broadcast}
}

func (s *slot) subscribe(topicID string, consumer *Consumer) {
	s.slotMut.RLock()
	s.topicMut.Lock()
	t, ok := s.topics[topicID]
	if !ok {
		t = newTopic(topicID, s.broadcast)
		s.topics[topicID] = t
	}
	s.topicMut.Unlock()
	t.subscribe(consumer)
	s.slotMut.RUnlock()
}

func (s *slot) unsubscribe(topicID string, consumerID string) {
	s.topicMut.RLock()
	t, ok := s.topics[topicID]
	s.topicMut.RUnlock()
	if ok {
		// t.unsubscribe(consumerID)
		if count := t.unsubscribe(consumerID); count <= 0 {
			s.slotMut.Lock()
			s.topicMut.Lock()
			if 0 >= t.consumerCount() {
				delete(s.topics, topicID)
			}
			s.topicMut.Unlock()
			s.slotMut.Unlock()
		}
	}
}

func (s *slot) publish(topicID string, from Session, message interface{}) {
	s.topicMut.RLock()
	t, ok := s.topics[topicID]
	s.topicMut.RUnlock()
	if ok {
		t.publish(from, message)
	}
}

func (s *slot) topicCount() uint64 {
	s.topicMut.RLock()
	defer s.topicMut.RUnlock()
	return uint64(len(s.topics))
}

func (s *slot) consumerCount() map[string]uint64 {
	s.topicMut.RLock()
	defer s.topicMut.RUnlock()
	count := map[string]uint64{}
	for _, topic := range s.topics {
		count[topic.id] = topic.consumerCount()
	}
	return count
}

type topic struct {
	id        string
	broadcast func(topicID string, to map[string]*Consumer, from Session, message interface{})
	consumers map[string]*Consumer
	mut       sync.RWMutex
}

func newTopic(id string, broadcast func(topicID string, to map[string]*Consumer, from Session, message interface{})) *topic {
	return &topic{id: id, consumers: make(map[string]*Consumer), broadcast: broadcast}
}

func (t *topic) consumerCount() uint64 {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return uint64(len(t.consumers))
}

func (t *topic) subscribe(consumer *Consumer) {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.consumers[consumer.GetSession().GetSessionImpl().GetID()] = consumer
}

func (t *topic) unsubscribe(consumerID string) (consumerCount int) {
	t.mut.Lock()
	defer t.mut.Unlock()
	delete(t.consumers, consumerID)
	return len(t.consumers)
}

func (t *topic) publish(from Session, message interface{}) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	t.broadcast(t.id, t.consumers, from, message)
}
