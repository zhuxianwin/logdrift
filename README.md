# logdrift

A CLI tool that tails and diffs structured JSON logs across multiple services in real time.

---

## Installation

```bash
go install github.com/yourusername/logdrift@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logdrift.git
cd logdrift
go build -o logdrift .
```

---

## Usage

Tail logs from multiple services and highlight structural differences as they arrive:

```bash
logdrift tail --services auth-service,payment-service --format json
```

Diff two log streams by field keys:

```bash
logdrift diff --left auth-service.log --right payment-service.log --key request_id
```

Watch live log output and flag when a service emits unexpected or missing fields:

```bash
logdrift watch --config logdrift.yaml
```

### Example Output

```
[auth-service]     {"level":"info","request_id":"abc123","latency_ms":42}
[payment-service]  {"level":"info","request_id":"abc123","latency_ms":98}
~ DRIFT DETECTED: latency_ms differs (42 vs 98)
```

---

## Configuration

`logdrift` can be configured via a `logdrift.yaml` file:

```yaml
services:
  - name: auth-service
    source: /var/log/auth/app.log
  - name: payment-service
    source: /var/log/payment/app.log
diff_keys:
  - request_id
  - user_id
```

---

## License

MIT © 2024 Your Name