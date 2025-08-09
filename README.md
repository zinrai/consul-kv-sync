# consul-kv-syncer

A tool to synchronize YAML configuration files to Consul KV store using the [Transaction API](https://developer.hashicorp.com/consul/api-docs/txn).

## Features

- Batch synchronization of multiple YAML files to Consul KV
- Environment-based configuration management
- Duplicate key detection across files
- Dry-run mode for validation
- Atomic operations using Consul Transaction API (up to 64 operations per transaction)

## Installation

```bash
$ go install github.com/zinrai/consul-kv-syncer
```

## Quick Start

Use the example directly from the example directory:

```bash
$ cd example
$ consul-kv-syncer -env development -dry-run
```

## Usage

Show help:

```bash
$ consul-kv-syncer -h
```

Sync production environment:

```bash
$ consul-kv-syncer -env production
```

Dry run for staging environment:

```bash
$ consul-kv-syncer -env staging -dry-run
```

Verbose output with custom Consul address:

```bash
$ consul-kv-syncer -env production -consul-addr http://consul:8500 -verbose
```

## Configuration

See the `example/` directory for sample configurations demonstrating:

- Environment-based organization
- Different configuration patterns
- Progressive complexity from development to production

## How it Works

1. Reads environment definition from `environments.yaml`
2. Loads all YAML files specified for the target environment
3. Detects duplicate keys across files
4. Converts nested YAML structure to flat key-value pairs
5. Synchronizes to Consul using Transaction API in batches

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
