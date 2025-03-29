## Table Schema & Partitioning 設計

本設計針對 **通知系統（notify-service）**，使用 **RANGE Partitioning** 來優化時間序列型資料的查詢與管理，避免單表過大影響效能。

---

### 1. Partition 設計概述

所有表格均以 `created_at` 進行 **RANGE Partitioning**，確保新舊資料有效分區：

- **優點**：
  - 加速查詢：透過 `created_at` 篩選時，僅掃描對應的 Partition，提高效能。
  - 簡化資料管理：可定期刪除舊 Partition，避免歷史資料影響效能。

| Database Name | Table Name        | Partition Key | 主要用途       |
| ------------- | ----------------- | ------------- | -------------- |
| `notify `     | `notify.messages` | `created_at`  | 訊息發送與排程 |
| `notify `     | `notify.queues`   | `created_at`  | 訊息發送隊列   |
| `notify `     | `notify.targets`  | `created_at`  | 訊息接收者紀錄 |
| `notify `     | `notify.events`   | `created_at`  | 訊息發送事件   |

---

### 2. 資料表結構

#### **2.1 `notify.messages`（通知訊息）**

**Partition Key**: `created_at`

- **用途**：記錄通知訊息，如 Email、SMS，支援排程發送。

```sql
CREATE TABLE notify.messages (
    id VARCHAR,
    type VARCHAR,
    sender_name VARCHAR,
    sender_address VARCHAR,
    subject VARCHAR,
    data TEXT,
    status VARCHAR,
    scheduled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_message_id ON notify.messages (id);

COMMENT ON COLUMN notify.messages.id              IS '流水號';
COMMENT ON COLUMN notify.messages.type            IS '消息類型 (sms/mail)';
COMMENT ON COLUMN notify.messages.sender_name     IS '郵件寄件者名稱';
COMMENT ON COLUMN notify.messages.sender_address  IS '郵件寄件者電子郵件地址';
COMMENT ON COLUMN notify.messages.subject         IS '郵件主旨';
COMMENT ON COLUMN notify.messages.data            IS '消息內容 (text/html)';
COMMENT ON COLUMN notify.messages.status          IS '狀態 (pending/enqueued/scheduled/canceled)';
COMMENT ON COLUMN notify.messages.scheduled_at    IS '預約時間';
COMMENT ON COLUMN notify.messages.created_at      IS '創建時間';
COMMENT ON COLUMN notify.messages.updated_at      IS '更新時間';
COMMENT ON COLUMN notify.messages.deleted_at      IS '刪除時間';
```

---

#### **2.2 `notify.queues`（發送隊列）**

**Partition Key**: `created_at`

- **用途**：管理訊息發送狀態，確保可靠傳遞。

```sql
CREATE TABLE notify.queues (
    id VARCHAR,
    status VARCHAR,
    message_id VARCHAR,
    driver VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_queue_id ON notify.queues (id);
CREATE INDEX idx_queue_status ON notify.queues (status);

COMMENT ON COLUMN notify.queues.id              IS '流水號';
COMMENT ON COLUMN notify.queues.status          IS '隊列狀態 (pending/enqueued/sending/success/failed)';
COMMENT ON COLUMN notify.queues.message_id      IS '對應的 messages.id';
COMMENT ON COLUMN notify.queues.driver          IS '發送驅動 (kafka)';
COMMENT ON COLUMN notify.queues.created_at      IS '創建時間';
COMMENT ON COLUMN notify.queues.updated_at      IS '更新時間';
COMMENT ON COLUMN notify.queues.deleted_at      IS '刪除時間';
```

---

#### **2.3 `notify.targets`（通知接收者）**

**Partition Key**: `created_at`

- **用途**：記錄每封通知的接收者資訊。

```sql
CREATE TABLE notify.targets (
    id VARCHAR,
    message_id       VARCHAR,
    receiver        VARCHAR,
    receiver_hash    VARCHAR,
    status          VARCHAR,
    queue_id         VARCHAR,
    driver_trace_id   VARCHAR,
    provider        VARCHAR,
    provider_trace_id VARCHAR,
    created_at       TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    deleted_at       TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_target_message_id ON notify.targets (message_id);
CREATE INDEX idx_target_receiver_hash ON notify.targets (receiver_hash);
CREATE INDEX idx_target_status ON notify.targets (status);
CREATE INDEX idx_target_queue_id ON notify.targets (queue_id);

COMMENT ON COLUMN notify.targets.id                IS '流水號';
COMMENT ON COLUMN notify.targets.message_id        IS '對應的 messages.id';
COMMENT ON COLUMN notify.targets.receiver          IS '接收者 (AES 加密)';
COMMENT ON COLUMN notify.targets.receiver_hash     IS '接收者 MD5 雜湊值';
COMMENT ON COLUMN notify.targets.status            IS '發送狀態 (pending/enqueued/sending/success/failed)';
COMMENT ON COLUMN notify.targets.queue_id          IS '對應的 queues.id';
COMMENT ON COLUMN notify.targets.driver_trace_id   IS '驅動器 trace_id';
COMMENT ON COLUMN notify.targets.provider          IS '訊息發送供應商';
COMMENT ON COLUMN notify.targets.provider_trace_id IS '供應商 trace_id';
COMMENT ON COLUMN notify.targets.created_at        IS '創建時間';
COMMENT ON COLUMN notify.targets.updated_at        IS '更新時間';
COMMENT ON COLUMN notify.targets.deleted_at        IS '刪除時間';
```

---

#### **2.4 `notify.events`（通知發送事件）**

**Partition Key**: `created_at`

- **用途**：記錄通知的發送狀態與回報。

```sql
CREATE TABLE notify.events (
    id                VARCHAR,
    provider          VARCHAR,
    status            VARCHAR,
    provider_trace_id VARCHAR,
    queue_id          VARCHAR,
    data              JSON,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL,
    deleted_at        TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX idx_events_status_provider_trace_id ON notify.events (status, provider_trace_id);

COMMENT ON COLUMN notify.events.id                IS '流水號';
COMMENT ON COLUMN notify.events.provider          IS '訊息發送供應商';
COMMENT ON COLUMN notify.events.status            IS '發送狀態 (sent/delivered/opened/clicked/failed)';
COMMENT ON COLUMN notify.events.provider_trace_id IS '供應商 trace_id';
COMMENT ON COLUMN notify.events.queue_id          IS '對應的 queues.id';
COMMENT ON COLUMN notify.events.data              IS '事件資料 (JSON)';
COMMENT ON COLUMN notify.events.created_at        IS '創建時間';
COMMENT ON COLUMN notify.events.updated_at        IS '更新時間';
COMMENT ON COLUMN notify.events.deleted_at        IS '刪除時間';
```

---

### 3. 設定 `pg_partman` 進行自動分區管理

> **自動創建每日 Partition**

### 3.1 安裝 `pg_partman`

```sql
-- 確保擁有 rds_superuser 權限
CREATE SCHEMA partman;
CREATE EXTENSION pg_partman WITH SCHEMA partman;
```

### 3.2 `create_parent` 函數配置分區

```sql
SELECT partman.create_parent(
    p_parent_table => 'notify.messages',
    p_control      => 'created_at',
    p_type         => 'range',
    p_interval     => '1 day',
    p_premake      => 7);

SELECT partman.create_parent(
    p_parent_table => 'notify.queues',
    p_control      => 'created_at',
    p_type         => 'range',
    p_interval     => '1 day',
    p_premake      => 7);

SELECT partman.create_parent(
    p_parent_table => 'notify.targets',
    p_control      => 'created_at',
    p_type         => 'range',
    p_interval     => '1 day',
    p_premake      => 7);

SELECT partman.create_parent(
    p_parent_table => 'notify.events',
    p_control      => 'created_at',
    p_type         => 'range',
    p_interval     => '1 day',
    p_premake      => 7);
```

### 3.3 設定 Partition Tables 的保留政策

```sql
UPDATE partman.part_config
SET infinite_time_partitions = true,
    retention = '90 days',
    retention_keep_table = true
WHERE parent_table = 'notify.messages';

UPDATE partman.part_config
SET infinite_time_partitions = true,
    retention = '90 days',
    retention_keep_table = true
WHERE parent_table = 'notify.queues';

UPDATE partman.part_config
SET infinite_time_partitions = true,
    retention = '90 days',
    retention_keep_table = true
WHERE parent_table = 'notify.targets';

UPDATE partman.part_config
SET infinite_time_partitions = true,
    retention = '90 days',
    retention_keep_table = true
WHERE parent_table = 'notify.events';
```

### 3.4 啟用 `pg_cron` 將分區維護作設置為自動運行

```sql
CREATE EXTENSION pg_cron;
SELECT cron.schedule_in_database('notify-partman-maintenance', '0 * * * *', $$CALL partman.run_maintenance_proc()$$, 'notify');
```

---

### 4. Partitioning 優勢

- **查詢效能提升**：

  - `created_at` 篩選條件會 **自動跳過無關 Partition**，避免全表掃描。
  - `INDEX` 只需維護 Partition 級別，避免單表索引過大。

- **資料管理簡化**：
  - 可以**快速刪除舊 Partition**，例如刪除 3 個月前的資料：
    ```sql
    DROP TABLE notify.messages_20241231;
    ```

---

### 5. 最佳化建議

1. **確保 `created_at` 為 `NOT NULL`**

   - `created_at` 為 Partition Key，若有 `NULL` 值則無法正確分區。

2. **確保 `created_at` `DEFAULT NOW()`**

   - 避免 `INSERT` 忘記帶 `created_at`，導致資料存入 **Default Partition（未分區）**。

---

### 6. 總結

- **Partition Key**：`created_at`，確保時間序列資料查詢效能。
- **索引設計**：針對 `id`、`status`、`queue_id` 等進行索引優化。
- **自動分區**：透過 `pg_partman` 創建每日 Partition，定期清理舊資料。

這樣的設計可有效應對高流量通知系統的需求，確保讀取與寫入效能最優化。🚀
