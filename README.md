# Event-Driven Job Queue

> **Crash-resilient background job system**  
> _Mini Sidekiq / Celery_

---

## Overview
> A persistent, **event-driven background job queue** designed to execute jobs reliably under crashes and shutdowns.
>
> _This **system guarantees at-least-once execution** and **explicitly does NOT attempt exactly-once semantics.**_
---

## Problem Statement
> Background job processing is deceptively hard.
>
> - Processes crash  
> - Workers die mid-execution  
> - Shutdowns happen at the worst possible time  
> - Losing jobs is unacceptable  
> - Preventing all duplicate execution is impractical  

The core problem is executing background jobs **reliably under failure**, without losing work, while keeping the system **simple, debuggable, and correct**.



---

## Guarantees

This system guarantees:

- **At-least-once execution**  
- **No job loss after persistence**  
- **Eventual recovery of stuck jobs via visibility timeouts**  
- **Bounded retries and bounded concurrency**  
- **Graceful shutdown without partial job state writes**

Duplicate execution is possible by design and must be handled via idempotent side effects where required.

This system does **NOT** guarantee:

- **Exactly-once execution**  
- **Distributed fault tolerance**  
- **Global job ordering**  
- **Real-time execution guarantees**

**These trade-offs are intentional and enable simpler recovery and failure handling.**

---

## Non-Goals

The system explicitly does **NOT** attempt to solve:

- **Exactly-once semantics**  
- **Distributed scheduling across nodes**  
- **High-throughput streaming**  
- **Horizontal database scalability**

**The design prioritizes correctness, clarity, and failure-mode reasoning over scale.**

---

## ⚙️ Architecture
**Execution flow:**
<p align="center">
  <img src="docs/architecture.png" alt="Event-Driven Job Queue Architecture" width="900">
</p>

**Core principles:**

- **The database is the single source of truth**
- **In-memory components coordinate execution, not correctness**
- **Correctness is enforced via atomic state transitions, not worker behavior**
- **Scheduling decisions are centralized to simplify correctness reasoning**



## Design Details

**Full design rationale, failure modes, and explicit trade-offs are documented here:**

👉 **[DESIGN.pdf](docs/DESIGN.pdf)**

## Build & Run

### Prerequisites

- **Go 1.20+**
- **No external dependencies**
  - SQLite is embedded via `modernc.org/sqlite`

---

### Compile

Build the server binary:

```bash
go build -o bin/server ./cmd/server 
```
This produces a standalone executable at: 
```bash
bin/server
```

### Run 

Start the job queue server:
```bash
./bin/server
```

### Submit a Job

Jobs are submitted via HTTP :
```bash
curl -X POST http://localhost:8080/createJob \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "payload": {
      "email": "user@example.com",
      "subject": "Welcome",
      "message": "Hello"
    },
    "max_retries": 3,
    "idempotency_key": "welcome-New-user-123"
  }'
```

### Response Semantics

- `201 Created`
Job was durably persisted and scheduled for execution.
- `429 Too Many Requests`
System is under backpressure. Client should retry later.






