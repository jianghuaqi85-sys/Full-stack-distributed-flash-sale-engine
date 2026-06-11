package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, key, value []byte) error

// Consumer Kafka 消费者封装（支持消费者组）
type Consumer struct {
	client       *kgo.Client
	topic        string
	group        string
	handler      MessageHandler
	maxRetries   int
	dlqTopic     string
	dlqProducer  *Producer
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers    []string
	Topic      string
	Group      string
	Handler    MessageHandler
	MaxRetries int    // 最大重试次数，超过则发送到 DLQ
	DLQTopic   string // 死信队列 topic
}

// NewConsumer 创建 Kafka 消费者
func NewConsumer(cfg ConsumerConfig, opts ...kgo.Opt) (*Consumer, error) {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	defaultOpts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumeTopics(cfg.Topic),
		kgo.ConsumerGroup(cfg.Group),
		kgo.DisableAutoCommit(),          // 手动提交，确保处理成功后才确认
		kgo.FetchMaxBytes(10 << 20),      // 10MB
		kgo.FetchMaxWait(500 * time.Millisecond),
		kgo.SessionTimeout(30 * time.Second),
		kgo.RebalanceTimeout(30 * time.Second),
	}
	defaultOpts = append(defaultOpts, opts...)

	client, err := kgo.NewClient(defaultOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	c := &Consumer{
		client:     client,
		topic:      cfg.Topic,
		group:      cfg.Group,
		handler:    cfg.Handler,
		maxRetries: cfg.MaxRetries,
		dlqTopic:   cfg.DLQTopic,
	}

	// 如果配置了死信队列，创建 DLQ producer
	if cfg.DLQTopic != "" {
		dlqProducer, err := NewProducer(cfg.Brokers, cfg.DLQTopic)
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
		}
		c.dlqProducer = dlqProducer
	}

	return c, nil
}

// Start 启动消费循环
func (c *Consumer) Start(ctx context.Context) {
	log.Printf("[Kafka] Consumer started: topic=%s, group=%s", c.topic, c.group)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka] Consumer stopping: topic=%s, group=%s", c.topic, c.group)
			c.client.Close()
			if c.dlqProducer != nil {
				c.dlqProducer.Close()
			}
			return
		default:
			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}

			fetches.EachError(func(topic string, partition int32, err error) {
				log.Printf("[Kafka] Fetch error: topic=%s partition=%d err=%v", topic, partition, err)
			})

			var records []*kgo.Record
			fetches.EachRecord(func(record *kgo.Record) {
				records = append(records, record)
			})

			for _, record := range records {
				c.processRecord(ctx, record)
			}

			// 批量提交 offset
			if len(records) > 0 {
				if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
					log.Printf("[Kafka] Failed to commit offsets: %v", err)
				}
			}
		}
	}
}

// processRecord 处理单条消息，带重试和死信队列
func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) {
	retryCount := getRetryCount(record)

	for i := retryCount; i < c.maxRetries; i++ {
		err := c.handler(ctx, record.Key, record.Value)
		if err == nil {
			return // 处理成功
		}

		log.Printf("[Kafka] Handler error (attempt %d/%d): %v", i+1, c.maxRetries, err)
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // 简单退避
	}

	// 超过最大重试次数，发送到死信队列
	if c.dlqProducer != nil {
		log.Printf("[Kafka] Message sent to DLQ: topic=%s offset=%d", c.topic, record.Offset)
		dlqRecord := &kgo.Record{
			Topic: c.dlqTopic,
			Key:   record.Key,
			Value: record.Value,
			Headers: []kgo.RecordHeader{
				{Key: "original-topic", Value: []byte(record.Topic)},
				{Key: "original-offset", Value: []byte(fmt.Sprintf("%d", record.Offset))},
				{Key: "retry-count", Value: []byte(fmt.Sprintf("%d", c.maxRetries))},
				{Key: "error-time", Value: []byte(time.Now().Format(time.RFC3339))},
			},
		}
		if err := c.dlqProducer.ProduceSync(ctx, dlqRecord.Key, dlqRecord.Value); err != nil {
			log.Printf("[Kafka] Failed to send to DLQ: %v", err)
		}
	} else {
		log.Printf("[Kafka] Message dropped (no DLQ configured): topic=%s offset=%d", c.topic, record.Offset)
	}
}

// getRetryCount 从消息头获取重试次数
func getRetryCount(record *kgo.Record) int {
	for _, header := range record.Headers {
		if header.Key == "retry-count" {
			var count int
			fmt.Sscanf(string(header.Value), "%d", &count)
			return count
		}
	}
	return 0
}
