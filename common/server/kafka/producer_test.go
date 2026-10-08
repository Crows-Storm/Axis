package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

type scriptedSender struct {
	delay time.Duration
	err   error
	sent  int
}

func (s *scriptedSender) SendMessages(msgs []*sarama.ProducerMessage) error {
	s.sent += len(msgs)
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	return s.err
}

func (s *scriptedSender) Close() error { return nil }

func TestSendDoesNotStartWhenContextIsCanceled(t *testing.T) {
	sender := &scriptedSender{}
	p := &Producer{sync: sender}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.Send(ctx, []*sarama.ProducerMessage{{Topic: "axis.user"}})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 0, sender.sent)
}

func TestSendReturnsProducerResultWhenContextEndsDuringSend(t *testing.T) {
	sender := &scriptedSender{delay: 30 * time.Millisecond}
	p := &Producer{sync: sender}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	err := p.Send(ctx, []*sarama.ProducerMessage{{Topic: "axis.user"}})
	require.NoError(t, err)
	require.Equal(t, 1, sender.sent)
}

func TestSendReturnsProducerError(t *testing.T) {
	p := &Producer{sync: &scriptedSender{err: errors.New("not enough replicas")}}
	err := p.Send(context.Background(), []*sarama.ProducerMessage{{Topic: "axis.user"}})
	require.ErrorContains(t, err, "not enough replicas")
}
