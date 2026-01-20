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
>
> Processes crash.
>
> Workers die mid-execution.
>
> Shutdowns happen at the worst possible time.
>
>
>Losing jobs is unacceptable.
>
>Preventing all duplicate execution is impractical.

---

## Guarantees

---

## Non-Goals

---

## ⚙️ Architecture

