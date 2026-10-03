CREATE TABLE outbox_messages (
                                 id              UUID PRIMARY KEY,
                                 aggregate_id    VARCHAR(255) NOT NULL,
                                 event_name      VARCHAR(255) NOT NULL,
                                 payload         JSONB NOT NULL,
                                 metadata        JSONB,
                                 status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                                 retry_count     INT NOT NULL DEFAULT 0,
                                 created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                 published_at    TIMESTAMPTZ,
                                 error_msg       TEXT
);

-- 查询 pending 消息的索引
CREATE INDEX idx_outbox_pending ON outbox_messages (status, created_at)
    WHERE status = 'pending';

-- 清理旧消息的索引
CREATE INDEX idx_outbox_created ON outbox_messages (created_at);

-- 幂等性表（消费端去重）
CREATE TABLE processed_events (
                                  event_id        UUID PRIMARY KEY,
                                  event_name      VARCHAR(255) NOT NULL,
                                  aggregate_id    VARCHAR(255) NOT NULL,
                                  processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 定期清理已处理事件的索引
CREATE INDEX idx_processed_events_time ON processed_events (processed_at);