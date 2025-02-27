# Workflow Engine

> Workflow engine for managing and executing workflow tasks.

## Structure

```text
.
├── context
│    ├── variables.go
│    └── context.go
├── batcher
│    ├── batcher.go
│    └── batcher_test.go
├── queue
│    └── queue.go
├── scheduler
│    ├── scheduler_test.go
│    └── scheduler.go
├── coordinator
│    ├── coordinator.go
│    └── coordinator_test.go
├── worker
│    ├── worker.go
│    └── worker_test.go
├── interface.go
├── errors.go
├── metrics
│    ├── example_test.go
│    ├── metrlcs_test.go
│    └── metrics.go
├── engine.go
├── executor
│    ├── types.go
│    ├── retry.go
│    ├── base.go
│    ├── manager.go
│    ├── process.go
│    ├── node.go
│    ├── service.go
│    └── task.go
├── handler
│    ├── subprocess.go
│    ├── notification.go
│    ├── approval.go
│    ├── service.go
│    ├── parallel.go
│    ├── script.go
│    ├── timer.go
│    ├── exclusive.go
│    ├── base.go
│    └── manager.go
├── concurrency.go
├── config
│    └── config.go
├── data_flow.go
├── state.go
├── core.go
└── README.md


```
