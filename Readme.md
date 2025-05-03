Overview

This messaging system is designed using the vertical slice architecture to ensure modularity and maintainability. It handles scheduled message delivery using PostgreSQL as the primary data store, Redis for caching and inter-process coordination, and includes mechanisms for failure handling and horizontal scalability.

Architecture Highlights

✅ Vertical Slice Architecture

Each use case is isolated in its own slice (e.g., sending, retrying, dead-letter handling), with minimal coupling.

✅ PostgreSQL

Row-level locking with FOR UPDATE SKIP LOCKED is used to safely fetch unsent messages in a concurrent environment.

Ensures that multiple workers can safely pull messages without double-processing.

✅ Redis

Caching: Sent messages are marked in Redis to avoid duplicate webhook invocations.

Pub/Sub: Used to remotely control the scheduler by broadcasting start/stop commands.

✅ Dead Letter Queue (DLQ) and Retry Mechanism

Messages that fail are inserted into a retry table with exponential backoff support.

After N failed attempts, messages are moved to a DLQ for manual inspection.

DLQ includes indexes for scheduled cleanup.

Scale-Out Strategy

1. Moderate Scale (Hundreds to Thousands of Messages per Hour)

This implementation effectively supports current requirements, with room to scale based on future load.

Vertical slices make it easy to isolate bottlenecks.

2. Higher Scale (Tens of Thousands per Hour)

PostgreSQL Partitioning:

Time-based (daily or hourly) partitions on the messages table.

Indexes on sent, created_at, and sent_at help maintain performance.

3. Enterprise Scale (Millions of Messages per Hour)

Replace polling with Kafka or Nats:

Use Kafka for durable, high-throughput ingestion.

PostgreSQL can still be the source of truth, but message queues allow real-time streaming.

Backpressure and Retry Topic:

Use separate topics for retries and dead-lettered messages.

4. Connection Optimization

Use PgBouncer for connection pooling to avoid overloading PostgreSQL with many short-lived connections.

Benefits

✔️ Simple and extensible.

✔️ Works well in Dockerized environments.

✔️ Capable of horizontal scaling without message duplication.

Future Enhancements

Metrics and observability via Prometheus/Grafana.

Admin UI for inspecting the DLQ and resending.

Optional webhook response validation and retry strategy configuration per customer.

Conclusion

This architecture offers a reliable, scalable, and maintainable solution for scheduled message delivery. It performs well in moderate-load scenarios and offers a clear upgrade path toward distributed streaming solutions like Kafka.

🚀 To run

```
docker-compose down -v --remove-orphans
docker-compose up --scale messaging-app=2 --build

docker-compose down -v --remove-orphans
docker-compose up --scale messaging-app=4 --build

docker-compose down -v --remove-orphans
docker-compose up --scale messaging-app=8 --build

and so on ..
```

scheduler can be configured with config.yaml

```
scheduler:
    enabled: true
    interval: 2m
    batchSize: 2
    maxConcurrent: 2
```

🌐 Swagger url
http://localhost:8080/swagger/index.html


