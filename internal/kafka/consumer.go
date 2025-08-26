package kafka

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	Brokers  []string
	GroupID  string
	Topics   []string
	ClientID string
}

func setDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

type MessageHandler func(ctx context.Context, rec *kgo.Record) error

func HandleLog() MessageHandler {
	return func(ctx context.Context, r *kgo.Record) error {
		log.Printf("DEBUG: topic=%s part=%d off=%d key=%q val=%q",
			r.Topic, r.Partition, r.Offset, r.Key, r.Value)
		return nil
	}
}

func RunConsumer(ctx context.Context, cfg Config, handle MessageHandler) error {
	if len(cfg.Brokers) == 0 || len(cfg.Topics) == 0 || cfg.GroupID == "" {
		return errors.New("kafka: empty brokers/topics/groupID")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topics...),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.ClientID(setDefault(cfg.ClientID, "web_demoservice-consumer")),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return err
	}
	defer cl.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()
	log.Printf("INFO: [kafka] started: brokers=%v group=%s topics=%v", cfg.Brokers, cfg.GroupID, cfg.Topics)

	for {
		fetches := cl.PollFetches(ctx)
		if err := ctx.Err(); err != nil {
			log.Printf("ERROR: [kafka] shutdown: %v", err)
			return nil
		}
		if fes := fetches.Errors(); fes != nil {
			for _, fe := range fes {
				log.Printf("ERROR: [kafka] fetch error: topic=%s part=%d err=%v", fe.Topic, fe.Partition, fe.Err)
			}
			continue
		}

		var toCommit []*kgo.Record
		fetches.EachRecord(
			func(r *kgo.Record) {
				if err := handle(ctx, r); err != nil {
					log.Printf("ERROR: [kafka] handle error (topic=%s part=%d off=%d): %v", r.Topic, r.Partition, r.Offset, err)
					return
				}
				toCommit = append(toCommit, r)
			})
		if len(toCommit) > 0 {
			commitCt, cancel := context.WithTimeout(ctx, 5*time.Second)
			_ = cl.CommitRecords(commitCt, toCommit...)
			cancel()
		}
	}
}
