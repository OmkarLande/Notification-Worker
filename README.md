# Notification Worker

A pluggable, database-driven notification worker built in Go. It polls for pending tasks, executes application-specific providers, generates channel payloads, and delivers notifications through Email, Discord, Slack, and WhatsApp — with built-in retry and failure tracking.

---

## 1. Getting Started

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.26+ |
| Docker & Docker Compose | Latest |
| PostgreSQL | 16 (included in Docker Compose) |

### Setup

```bash
git clone https://github.com/OmkarLande/Notification-Worker.git
cd Notification-Worker

# Create your environment file
cp .env.example .env
```

### Start with Docker

```bash
# Start PostgreSQL, Redis, and the Worker
docker compose up --build

# Or run only the infrastructure (to develop locally)
docker compose up postgres redis
```

### Run Locally (without Docker)

```bash
# Ensure PostgreSQL is running and DB_URL in .env points to it

# Apply database migrations (requires golang-migrate CLI)
migrate -path internal/database/migrations -database "$DB_URL" up

# Run the worker
go run cmd/worker/main.go
```

### Useful Commands

```bash
# Build the binary
go build -o bin/notification-worker ./cmd/worker/main.go

# Run all tests
go test ./...

# Apply migrations manually
migrate -path internal/database/migrations -database "$DB_URL" up

# Roll back last migration
migrate -path internal/database/migrations -database "$DB_URL" down 1

# View container logs
docker compose logs -f app
```

> **Note:** Migrations are not auto-applied by the binary. You must run them manually before first use.

---

## 2. Architecture Overview

### High-Level Architecture

```mermaid
flowchart TD
    subgraph Trigger
        A["n8n / Scheduler"]
    end
    
    subgraph Storage
        B["Jobs Table"]
        D["Task Queue (DB)"]
    end
    
    subgraph Dispatch
        C["Task Creation"]
        E["Task Dispatcher"]
    end
    
    subgraph Worker
        F["Task Worker"]
        
        subgraph Processing Pipeline
            G["Execution Pipeline"]
            H["Provider"]
            I["Content Generation"]
            J["Channel Delivery"]
            K["Reliability Layer"]
        end
    end
    
    L["Completed / Retry"]

    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    H --> I
    I --> J
    J --> K
    K --> L
```

### Execution Pipeline

Every task flows through a sequential pipeline of six steps:

```mermaid
flowchart TD
    A["ValidateContextStep (10)"] --> B["ProviderExecutionStep (20)"]
    B --> C["InsightGenerationStep (30)"]
    C --> D["PayloadTransformationStep (40)"]
    D --> E["ChannelDeliveryStep (50)"]
    E --> F["FinalizeExecutionStep (100)"]

    B -- "on failure" --> F
    C -- "on failure" --> F
    D -- "on failure" --> F
    E -- "on failure" --> F
```

If any step fails, the pipeline skips remaining steps and jumps directly to `FinalizeExecutionStep`, which invokes the Reliability Manager for logging, metrics, and retry scheduling.

### Database Overview

```mermaid
erDiagram
    apps ||--o{ jobs : "owns"
    jobs ||--o{ tasks : "creates"
    jobs ||--o{ job_channels : "has"
    channels ||--o{ job_channels : "assigned to"
    tasks ||--o{ channel_tasks : "delivers via"
    channels ||--o{ channel_tasks : "used by"
    tasks ||--o{ task_logs : "failure history"
    tasks ||--o| tasks : "retry lineage (parent_task_id)"
    job_status ||--o{ jobs : "status"
    task_status ||--o{ tasks : "status"
    channel_task_status ||--o{ channel_tasks : "status"
```

| Table | Purpose |
|---|---|
| **apps** | Registered applications with connection info and maintenance mode |
| **jobs** | Defines recurring notification work — retry limits, thread counts, default arguments |
| **tasks** | Runtime execution units created per user per job trigger |
| **channels** | Supported notification transports (Email, Discord, Slack, WhatsApp) |
| **job_channels** | Maps which channels a job should deliver to |
| **channel_tasks** | Maps runtime tasks to their delivery channels |
| **task_logs** | Failure history with performance and error payloads |
| **job_status** | Active, Disabled, Archived |
| **task_status** | NeedToPick, Picked, Processing, Completed, Failed, RetryScheduled, Cancelled |
| **channel_task_status** | Pending, Sent, Failed, RetryScheduled |

---

## 3. Registering a New Application

This section walks through onboarding a new application using the **Expense Tracker** as an example.

### Step 1 — Register the Application

Insert a row into the `apps` table:

```sql
INSERT INTO apps (name, base_url, connection_string, database_name, maintenance_mode)
VALUES (
    'expense',
    'https://expense.ogom.in',
    'postgresql://user:pass@host:5432/expense_db?sslmode=require',
    'expense_db',
    false
);
```

The `name` field must match the provider registry key used in `application.go`:

```go
factory.Register("expense", expenseProvider)
```

### Step 2 — Register Jobs

Each job represents a type of recurring notification. Insert into `jobs`:

```sql
-- Daily Digest
INSERT INTO jobs (app_id, name, description, status_id, max_thread_count, max_retry_count, arguments)
VALUES (
    (SELECT id FROM apps WHERE name = 'expense'),
    'Daily Digest',
    'Sends a daily summary of expenses to each user',
    (SELECT id FROM job_status WHERE name = 'Active'),
    4,
    3,
    '{"isDetailed": true}'
);

-- Monthly Summary
INSERT INTO jobs (app_id, name, description, status_id, max_thread_count, max_retry_count, arguments)
VALUES (
    (SELECT id FROM apps WHERE name = 'expense'),
    'Monthly Summary',
    'End-of-month expense report',
    (SELECT id FROM job_status WHERE name = 'Active'),
    2,
    3,
    '{"includeCharts": true}'
);
```

The `arguments` column holds **default** job arguments. Runtime tasks may override these.

### Step 3 — Assign Channels to the Job

Map which channels each job should deliver notifications through:

```sql
INSERT INTO job_channels (job_id, channel_id)
VALUES
    ((SELECT id FROM jobs WHERE name = 'Daily Digest'), (SELECT id FROM channels WHERE name = 'Email')),
    ((SELECT id FROM jobs WHERE name = 'Daily Digest'), (SELECT id FROM channels WHERE name = 'Discord'));
```

### Step 4 — Trigger from Scheduler

Jobs do **not** execute directly. A scheduler (e.g. n8n) triggers job execution by calling `JobExecutionService.TriggerJob(jobID)`. This:

1. Validates the job and app (checks active status, maintenance mode)
2. Calls the provider's `GetNotificationEnabledUsers()` to get the user list
3. Creates one **Task** per user with `status = NeedToPick`
4. Creates **channel_tasks** rows for each task × channel combination

```mermaid
flowchart LR
    A["n8n Cron"] --> B["TriggerJob(jobID)"]
    B --> C["Validate Job + App"]
    C --> D["Provider.GetUsers()"]
    D --> E["Create Tasks"]
    E --> F["Create ChannelTasks"]
    F --> G["Tasks wait in queue"]
```

The worker then polls for `NeedToPick` tasks via `TaskDispatcher.Run()`.

### Step 5 — Runtime Task Execution

Each task carries its own `arguments` JSON (e.g. `{"user_id": 42}`). The execution pipeline:

1. **ValidateContextStep** — Ensures Task, Job, App, Provider are non-nil
2. **ProviderExecutionStep** — Calls `provider.Execute(ctx, ec)` with the task arguments
3. **InsightGenerationStep** — Generates rule-based insights from the provider output
4. **PayloadTransformationStep** — Renders templates into Email HTML, Discord text, Slack blocks, WhatsApp text
5. **ChannelDeliveryStep** — Delivers payloads to each assigned channel via the channel registry
6. **FinalizeExecutionStep** — Records metrics, logs failures, and schedules retries via the Reliability Manager

Channel-specific configuration (Discord webhook URLs, email recipients) is resolved from the payload and channel configuration — **nothing is hardcoded in the worker**.

### Step 6 — Lifecycle & Retry

```mermaid
stateDiagram-v2
    [*] --> NeedToPick: Task Created
    NeedToPick --> Picked: Dispatcher assigns
    Picked --> Processing: Worker starts
    Processing --> Completed: All steps pass
    Processing --> Failed: Step error
    Failed --> NeedToPick: Retry scheduled (new task)

    note right of Failed
        Retry creates a NEW task with:
        - Same arguments
        - Same channels
        - parent_task_id → original
        - current_retry_count + 1
        Failed task is never mutated.
    end note
```

Retries preserve complete execution lineage:
- Every retry is a **new Task row** with `parent_task_id` pointing to the original
- `current_retry_count` is incremented
- All `channel_tasks` from the original are re-created for the retry task
- Retries stop when `current_retry_count >= job.max_retry_count`
- Failed tasks are **never mutated** — they remain as permanent audit records
