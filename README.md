# RoadmapKV

RoadmapKV is a systems-engineering exploration built to understand how a small
key-value service evolves from a local core into a networked system. The name
reflects that roadmap: each stage adds a systems concern that can be measured,
tested, and revisited rather than hidden behind a framework.

## Roadmap

Completed:

- Core in-memory engine
- TCP server and persistent client sessions
- Simple command parser and newline-delimited protocol
- Unit, integration, and race-tested behavior
- Interactive CLI client
- Initial parser and storage benchmark baselines

Planned:

- TTL and expiration
- Persistence, compaction, and reconstruction
- Measured optimizations
- Replication
- Raft exploration

## Design highlights

- TCP is treated as a byte stream: commands may be fragmented across reads or
  pipelined in one write.
- The server buffers incomplete commands and preserves request order.
- Storage uses synchronized access and returns copies of stored values.
- The client owns a persistent connection and handles request/response framing.
- Tests use ephemeral loopback connections and `net.Pipe` where each provides
  the clearest coverage.

## Setup

Requires Go 1.27.1 or later.

Start the server:

```bash
go run ./src
```

In another terminal, start the CLI:

```bash
go run ./src/cmd/client
```

The server listens on port `8080` by default. Set `PORT` to use another port:

```bash
PORT=9090 go run ./src
PORT=9090 go run ./src/cmd/client
```

Supported commands:

```text
SET key value
GET key
DEL key
PING
QUIT
```

Run the test suite:

```bash
go test -race ./...
```

Run benchmark baselines:

```bash
go test -bench=. -benchmem ./...
```