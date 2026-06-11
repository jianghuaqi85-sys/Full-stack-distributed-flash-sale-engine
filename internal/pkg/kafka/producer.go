package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer Kafka 生产者封装
type Producer struct {
	client *kgo.Client
	topic  string
}

// NewProducer 创建 Kafka 生产者
func NewProducer(brokers []string, topic string, opts ...kgo.Opt) (*Producer, error) {
	defaultOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
		kgo.RequiredAcks(kgo.AllISRAcks()), // 确保所有副本确认
		kgo.ProducerLinger(5 * time.Millisecond), // 批量发送延迟，提升吞吐
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)), // 相同 key 路由到同一 partition
	}
	defaultOpts = append(defaultOpts, opts...)

	client, err := kgo.NewClient(defaultOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		client: client,
		topic:  topic,
	}, nil
}

// ProduceSync 同步发送消息（等待确认）
func (p *Producer) ProduceSync(ctx context.Context, key, value []byte) error {
	record := &kgo.Record{
		Topic: p.topic,
		Key:   key,
		Value: value,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results := p.client.ProduceSync(ctx, record)
	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("failed to produce message: %w", result.Err)
		}
	}
	return nil
}

// ProduceAsync 异步发送消息（高性能路径）
func (p *Producer) ProduceAsync(ctx context.Context, key, value []byte, cb func(error)) {
	record := &kgo.Record{
		Topic: p.topic,
		Key:   key,
		Value: value,
	}

	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if cb != nil {
			cb(err)
		}
		if err != nil {
			log.Printf("[Kafka] Async produce error: %v", err)
		}
	})
}

// Close 关闭生产者
func (p *Producer) Close() {
	p.client.Close()
}
