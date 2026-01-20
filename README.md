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
> • Processes crash  
> • Workers die mid-execution  
> • Shutdowns happen at the worst possible time  
> • Losing jobs is unacceptable  
> • Preventing all duplicate execution is impractical


---

## Guarantees

This system guarantees:

**• At-least-once execution**  
**• No job loss after persistence**  
**• Eventual recovery of stuck jobs**  
**• Bounded retries and bounded concurrency**  
**• Graceful shutdown without partial writes**

This system does **NOT** guarantee:

**• Exactly-once execution**  
**• Distributed fault tolerance**  
**• Global job ordering**  
**• Real-time execution guarantees**

**These trade-offs are intentional and enable simpler recovery and failure handling.**

---

## Non-Goals

The system explicitly does **NOT** attempt to solve:

**• Exactly-once semantics**  
**• Distributed scheduling across nodes**  
**• High-throughput streaming**  
**• Horizontal database scalability**

**The design prioritizes correctness, clarity, and failure-mode reasoning over scale.**

---

## ⚙️ Architecture
**Execution flow:**
<p align="center">
  <img src="docs/architecture.png" alt="Event-Driven Job Queue Architecture" width="900">
</p>
**Core principles:**

**• The database is the single source of truth**  
**• In-memory components coordinate execution, not correctness**  
**• Correctness is enforced via atomic state transitions, not worker behavior**

