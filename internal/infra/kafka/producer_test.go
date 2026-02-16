package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/twmb/franz-go/pkg/kgo"
)

type fakeProducer struct {
	produced []*kgo.Record
	err      error
}

func (f *fakeProducer) Produce(ctx context.Context, record *kgo.Record, fn func(*kgo.Record, error)) {
	f.produced = append(f.produced, record)
	fn(record, f.err)
}

func TestProducer_Publish_ValidatesInput(t *testing.T) {
	p, err := newProducerWithClient(&fakeProducer{}, "dlq")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Publish(context.Background(), nil, errors.New("boom")); err == nil {
		t.Fatalf("expected error for nil source record")
	}
	if err := p.Publish(context.Background(), &kgo.Record{}, nil); err == nil {
		t.Fatalf("expected error for nil cause")
	}
}

func TestProducer_Publish_PropagatesProduceError(t *testing.T) {
	wantErr := errors.New("produce failed")
	fp := &fakeProducer{err: wantErr}
	p, err := newProducerWithClient(fp, "dlq")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = p.Publish(context.Background(), &kgo.Record{Topic: "src"}, errors.New("boom"))
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestProducer_Publish_AddsDLQHeaders(t *testing.T) {
	fp := &fakeProducer{}
	p, err := newProducerWithClient(fp, "dlq")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	src := &kgo.Record{
		Topic:     "source",
		Partition: 3,
		Offset:    12,
		Key:       []byte("k"),
		Value:     []byte("v"),
		Headers: []kgo.RecordHeader{
			{Key: "existing", Value: []byte("1")},
		},
	}

	if err := p.Publish(context.Background(), src, errors.New("boom")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fp.produced) != 1 {
		t.Fatalf("expected 1 produced record, got %d", len(fp.produced))
	}

	got := fp.produced[0]
	if got.Topic != "dlq" {
		t.Fatalf("expected dlq topic, got %s", got.Topic)
	}
	if string(got.Key) != string(src.Key) || string(got.Value) != string(src.Value) {
		t.Fatalf("expected key/value copied from source")
	}

	headers := make(map[string]string, len(got.Headers))
	for _, h := range got.Headers {
		headers[h.Key] = string(h.Value)
	}

	if headers["existing"] != "1" {
		t.Fatalf("expected existing header preserved")
	}
	if headers["dlq_error"] != "boom" {
		t.Fatalf("expected dlq_error header, got %q", headers["dlq_error"])
	}
	if headers["dlq_source_topic"] != "source" {
		t.Fatalf("expected dlq_source_topic header, got %q", headers["dlq_source_topic"])
	}
	if headers["dlq_source_partition"] != "3" {
		t.Fatalf("expected dlq_source_partition header, got %q", headers["dlq_source_partition"])
	}
	if headers["dlq_source_offset"] != "12" {
		t.Fatalf("expected dlq_source_offset header, got %q", headers["dlq_source_offset"])
	}
	if headers["dlq_ts"] == "" {
		t.Fatalf("expected dlq_ts header to be set")
	}
}
