<div align="center">
<h1>cdc-trigger</h1>

[![Auth](https://img.shields.io/badge/Auth-chihqiang-ff69b4)](https://github.com/chihqiang)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/pulls)
[![Release](https://img.shields.io/github/release/chihqiang/cdc-trigger.svg?style=flat-square)](https://github.com/chihqiang/cdc-trigger/releases)
[![GitHub Pull Requests](https://img.shields.io/github/stars/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/stargazers)
[![HitCount](https://views.whatilearened.today/views/github/chihqiang/cdc-trigger.svg)](https://github.com/chihqiang/cdc-trigger)
[![GitHub license](https://img.shields.io/github/license/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/blob/main/LICENSE)

<p>
 cdc-trigger is an efficient Go-based Change Data Capture (CDC) tool that real-time monitors database changes, parses and processes events, and sends them to message queues or other downstream systems.
</p>

</div>

## Project Overview

## Features

- **Real-time Capture**: Monitor database change events in real-time through binlog parsing
- **Unified Event Format**: Convert changes from different databases into a consistent JSON format
- **Multiple Output Support**: Send events to various downstream systems including stdout, Redis, Kafka, RabbitMQ, RocketMQ, and Pulsar
- **Checkpoint Resumption**: Store synchronization positions to achieve breakpoint resumption
- **Extensible Architecture**: Easy to extend with new data sources and output types
- **Worker Pool Processing**: Process events efficiently with worker goroutines
- **Graceful Shutdown**: Properly handle context cancellation and resource cleanup
- **Single Source Configuration**: One YAML file per deployment, with `${VAR}` / `${VAR:-fallback}` environment substitution inside it

## Supported Components

### Data Sources

- MySQL (via binlog parsing)

### Outputs

- Standard Output (stdout)
- [Redis](https://redis.io/)
- [Kafka](https://kafka.apache.org/)
- [RabbitMQ](https://www.rabbitmq.com/)
- [RocketMQ](https://rocketmq.apache.org/)
- [Pulsar](https://pulsar.apache.org/)

### Storage

- File Storage
- Redis Storage

## Quick Start

### Prerequisites

- Go 1.23+ environment
- MySQL server with binary logging enabled
- Correct database access permissions (MySQL user needs binlog read permissions)
- If using other output components, ensure corresponding services are available

### Installation

### Prerequisites

- Go **1.23+** (latest version recommended)
- `$GOPATH/bin` added to your `$PATH`

### Option 1: Install via `go install` (Recommended)

This is the simplest and recommended way to install **cdc-trigger**:

```bash
go install github.com/chihqiang/cdc-trigger/cmd/cdc-trigger@latest
```

### Option 2: Build from Source

If you want to modify the source code or contribute to development, build from source:

```bash
# Clone the repository
git clone https://github.com/chihqiang/cdc-trigger.git
cd cdc-trigger && make build
cp ./cdc-trigger /usr/local/bin/
```

### Usage Example

1. Create configuration file `config.yml`: (see Configuration File Description section for details)

2. Run the program:

```bash
# Using the default config.yml
cdc-trigger
# Using a specific config file
cdc-trigger -c path/to/config.yml
# Explicitly using the listen command
cdc-trigger listen -c path/to/config.yml
```

> The configuration file is required: the file is the only source of configuration, and starting
> without a readable file fails with an error. Values that differ per environment are not passed as
> environment variables directly, but referenced from inside the file — see
> [Environment Variables](#environment-variables).

## Configuration File Description

The configuration file uses YAML format and consists of three main parts: `store` (offset storage), `source` (data source), and `output` (output destination).

### Example Configuration

```yaml
# ==========================================
# cdc-trigger Configuration File Example (YAML)
# ==========================================
# The full annotated file is available at ./config.yml in the repository root.

# ---------- Offset Storage Configuration ----------
store:
  type: "${STORE_TYPE:-file}"              # Storage type: file / redis

  file:
    dir: "${STORE_FILE_DIR:-runtime}"      # Directory for storing offset files

  redis:
    addr: "${STORE_REDIS_ADDR:-127.0.0.1:6379}"   # Redis address
    password: "${STORE_REDIS_PASSWORD:-}"         # Redis password (leave empty if none)
    db: "${STORE_REDIS_DB:-0}"                    # Redis database number (default 0)

# ---------- Data Source Configuration ----------
source:
  type: "${SOURCE_TYPE:-mysql}"            # Data source type: mysql

  mysql:
    addr: "${SOURCE_MYSQL_ADDR:-127.0.0.1:3306}"  # Database address (host:port)
    user: "${SOURCE_MYSQL_USER:-root}"            # Database username (recommended to use a dedicated account in production)
    password: "${SOURCE_MYSQL_PASSWORD:-123456}"  # Database password
    exclude_table_regex:              # Tables to exclude (regex patterns)
      - "mysql.*"
      - "information_schema.*"
      - "performance_schema.*"
      - "sys.*"
    include_table_regex:              # Tables to include (regex patterns, empty = all except excluded)
      # - "cdctrigger.*"              # Example: only listen to cdctrigger tables

# ---------- Output Configuration ----------
output:
  type: "${OUTPUT_TYPE:-stdout}"       # Output type: stdout / kafka / redis / rabbitmq / rocketmq / pulsar

  # Kafka settings
  kafka:
    brokers:
      - "${OUTPUT_KAFKA_BROKERS:-127.0.0.1:9092}"  # Kafka broker list
    topic: "${OUTPUT_KAFKA_TOPIC:-cdc-trigger-events}"  # Kafka topic name

  # RabbitMQ settings
  rabbitmq:
    url: "${OUTPUT_RABBITMQ_URL:-amqp://guest:guest@127.0.0.1:5672/}" # RabbitMQ connection URL
    exchange: "${OUTPUT_RABBITMQ_EXCHANGE:-cdc-trigger-exchange}" # Exchange name (declared and bound on startup)
    exchange_type: "${OUTPUT_RABBITMQ_EXCHANGE_TYPE:-direct}"     # Exchange type: direct / fanout / topic (headers is refused)
    routing_key: "${OUTPUT_RABBITMQ_ROUTING_KEY:-cdc-trigger-events}" # Publish key, and the key the queue is bound with
    queue: "${OUTPUT_RABBITMQ_QUEUE:-cdc-trigger-events}"      # Queue name
    durable: true              # Whether the queue should survive server restarts (also makes messages persistent)
    auto_delete: false         # Whether the queue should auto-delete when unused
    auto_ack: false            # Whether to auto-acknowledge messages (unused when publishing)
    exclusive: false           # Whether the queue is exclusive to this connection
    no_wait: false             # Whether to wait for the server to confirm queue declaration

  # Redis settings
  redis:
    addr: "${OUTPUT_REDIS_ADDR:-127.0.0.1:6379}"        # Redis address
    password: "${OUTPUT_REDIS_PASSWORD:-}"              # Redis password
    db: "${OUTPUT_REDIS_DB:-0}"                         # Redis database number
    key: "${OUTPUT_REDIS_KEY:-cdc-trigger-events}"      # Redis key for storing events

  # RocketMQ settings
  rocketmq:
    servers:
      - "${OUTPUT_ROCKETMQ_SERVERS:-127.0.0.1:9876}"  # RocketMQ NameServer address
    topic: "${OUTPUT_ROCKETMQ_TOPIC:-cdc-trigger-events}"  # RocketMQ topic name
    group: "${OUTPUT_ROCKETMQ_GROUP:-cdc-trigger-group}"   # Producer group name
    namespace: "${OUTPUT_ROCKETMQ_NAMESPACE:-}"        # Namespace
    access_key: "${OUTPUT_ROCKETMQ_ACCESS_KEY:-}"      # Access key
    secret_key: "${OUTPUT_ROCKETMQ_SECRET_KEY:-}"      # Secret key
    retry: 3                   # Retry count on failure

  # Pulsar settings
  pulsar:
    url: "${OUTPUT_PULSAR_URL:-pulsar://127.0.0.1:6650}"  # Pulsar broker URL
    topic: "${OUTPUT_PULSAR_TOPIC:-cdc-trigger-events}"   # Pulsar topic name
    token: "${OUTPUT_PULSAR_TOKEN:-}"        # Optional authentication token
    operation_timeout: 30           # Operation timeout in seconds
    connection_timeout: 30          # Connection timeout in seconds
```

### Environment Variables

The file is the only source of configuration — there is no environment variable fallback for the
settings themselves. Environment variables are referenced **from inside the file** with `${VAR}`
or `${VAR:-fallback}`, which is how one file serves a laptop, a container and Kubernetes:

```yaml
source:
  mysql:
    addr: "${SOURCE_MYSQL_ADDR:-127.0.0.1:3306}"  # overridable, with a fallback
    user: "${SOURCE_MYSQL_USER:-root}"
    password: "${SOURCE_MYSQL_PASSWORD}"          # required from the environment
```

```bash
SOURCE_MYSQL_ADDR=mysql:3306 SOURCE_MYSQL_PASSWORD=secret ./cdc-trigger -c config.yml
```

| Syntax | Meaning |
| --- | --- |
| `${VAR}` | The value of `VAR`, or an empty string when it is unset |
| `${VAR:-fallback}` | The value of `VAR`, or `fallback` when `VAR` is unset or empty |
| `$$` | An escaped `$` (use it for a literal `${...}`) |

A key that is missing from the file falls back to the default written in the tag (`default=`), and a
key that is required but missing from both the file and the defaults fails at startup with the field
path in the error.

> **Note:** Expansion only ever fills a string or a key, so a value that should be a list — e.g.
> `exclude_table_regex` — has to be written as a list in the file; it cannot be passed as a
> comma-separated environment variable.

## Docker Deployment

You can use Docker to run cdc-trigger in containerized environments. Here's how to build and run cdc-trigger with Docker:

The image ships `/app/config.yml` (the file from the repository root) and is started with
`cdc-trigger -c /app/config.yml`. Every `-e` below fills a `${VAR:-fallback}` reference in that file,
so the same image can be pointed at another database, storage or output without a rebuild:

```bash
# =========================
# 1️⃣ MySQL only (read from MySQL)
# =========================
docker run -it --rm \
    --name cdc-trigger \
    -e SOURCE_MYSQL_ADDR="127.0.0.1:3306" \
    -e SOURCE_MYSQL_USER="root" \
    -e SOURCE_MYSQL_PASSWORD="123456" \
    zhiqiangwang/app:cdc-trigger

# =========================
# 2️⃣ MySQL → Redis & Redis
# =========================
docker run -it --rm \
    --name cdc-trigger \
    -e SOURCE_MYSQL_ADDR="127.0.0.1:3306" \
    -e SOURCE_MYSQL_USER="root" \
    -e SOURCE_MYSQL_PASSWORD="123456" \
    -e STORE_TYPE="redis" \
    -e STORE_REDIS_ADDR="127.0.0.1:6379" \
    -e STORE_REDIS_PASSWORD="123456" \
    -e STORE_REDIS_DB="1" \
    -e OUTPUT_TYPE="redis" \
    -e OUTPUT_REDIS_ADDR="127.0.0.1:6379" \
    -e OUTPUT_REDIS_PASSWORD="123456" \
    -e OUTPUT_REDIS_DB="1" \
    -e OUTPUT_REDIS_KEY="cdc-trigger-events" \
    zhiqiangwang/app:cdc-trigger
```

To use your own configuration file instead, mount it over the bundled one:

```bash
docker run -it --rm \
    --name cdc-trigger \
    -v "$PWD/config.yml:/app/config.yml:ro" \
    zhiqiangwang/app:cdc-trigger
```

The environment variables the bundled `config.yml` reads are `STORE_TYPE`, `STORE_FILE_DIR`,
`STORE_REDIS_ADDR`, `STORE_REDIS_PASSWORD`, `STORE_REDIS_DB`, `SOURCE_TYPE`, `SOURCE_MYSQL_ADDR`,
`SOURCE_MYSQL_USER`, `SOURCE_MYSQL_PASSWORD`, `OUTPUT_TYPE`, `OUTPUT_KAFKA_BROKERS`,
`OUTPUT_KAFKA_TOPIC`, `OUTPUT_RABBITMQ_URL`, `OUTPUT_RABBITMQ_EXCHANGE`, `OUTPUT_RABBITMQ_EXCHANGE_TYPE`,
`OUTPUT_RABBITMQ_ROUTING_KEY`, `OUTPUT_RABBITMQ_QUEUE`,
`OUTPUT_REDIS_ADDR`, `OUTPUT_REDIS_PASSWORD`, `OUTPUT_REDIS_DB`, `OUTPUT_REDIS_KEY`,
`OUTPUT_ROCKETMQ_SERVERS`, `OUTPUT_ROCKETMQ_TOPIC`, `OUTPUT_ROCKETMQ_GROUP`,
`OUTPUT_ROCKETMQ_NAMESPACE`, `OUTPUT_ROCKETMQ_ACCESS_KEY`, `OUTPUT_ROCKETMQ_SECRET_KEY`,
`OUTPUT_PULSAR_URL`, `OUTPUT_PULSAR_TOPIC` and `OUTPUT_PULSAR_TOKEN`.

## Notes

1. **MySQL Configuration Requirements**:
   - Binary logging must be enabled (`log-bin=ON`)
   - Server ID must be set (`server-id=1`)
   - Binlog format should be `ROW` (`binlog_format=ROW`)
2. **Permission Requirements**: When using MySQL data source, ensure the database user has sufficient permissions:

```sql
-- Create an account
CREATE USER 'cdctrigger'@'%' IDENTIFIED BY 'strong_password';

-- Authorization (the REPLICATION permission is required to read the binlog)
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdctrigger'@'%';

-- If cdc-trigger needs to do metadata queries, it also needs read permissions
GRANT SELECT ON *.* TO 'cdctrigger'@'%';

-- Refresh permissions
FLUSH PRIVILEGES;
```
