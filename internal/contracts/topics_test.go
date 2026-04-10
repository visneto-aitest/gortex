package contracts

import (
	"testing"
)

func TestTopicExtractor_GoPublisher(t *testing.T) {
	ext := &TopicExtractor{}
	src := []byte(`
func publish(p *kafka.Producer) {
	p.Produce("user.created", payload)
	nc.Publish("order.shipped", data)
}
`)
	contracts := ext.Extract("publisher.go", src, nil, nil)
	if len(contracts) != 2 {
		t.Fatalf("expected 2 provider contracts, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "topic::user.created", ContractTopic, RoleProvider)
	assertContract(t, contracts[1], "topic::order.shipped", ContractTopic, RoleProvider)
}

func TestTopicExtractor_GoSubscriber(t *testing.T) {
	ext := &TopicExtractor{}
	src := []byte(`
func consume(c *kafka.Consumer) {
	c.Subscribe("user.created", nil)
}
`)
	contracts := ext.Extract("consumer.go", src, nil, nil)
	if len(contracts) != 1 {
		t.Fatalf("expected 1 consumer contract, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "topic::user.created", ContractTopic, RoleConsumer)
}

func TestTopicExtractor_TSProducer(t *testing.T) {
	ext := &TopicExtractor{}
	src := []byte(`
await producer.send({ topic: "notifications", messages: [msg] });
`)
	contracts := ext.Extract("producer.ts", src, nil, nil)
	if len(contracts) != 1 {
		t.Fatalf("expected 1 provider contract, got %d", len(contracts))
	}
	assertContract(t, contracts[0], "topic::notifications", ContractTopic, RoleProvider)
}
