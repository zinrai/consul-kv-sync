# consul-kv-sync

A tool to synchronize YAML configuration files to Consul KV store using the [Transaction API](https://developer.hashicorp.com/consul/api-docs/txn).

## Features

- Batch synchronization of every YAML file in a directory to Consul KV
- Duplicate key detection across files
- Dry-run mode for validation
- Atomic operations using Consul Transaction API (up to 64 operations per transaction)
- Export to Consul KV JSON format for backup or import

## Quick Start

Use the example directly from the example directory:

```bash
$ cd example
$ consul-kv-sync -dir kv-files/development -dry-run
```

## Usage

Show help:

```bash
$ consul-kv-sync -h
```

Sync a directory:

```bash
$ consul-kv-sync -dir kv-files/production
```

Dry run:

```bash
$ consul-kv-sync -dir kv-files/staging -dry-run
```

Verbose output with custom Consul address:

```bash
$ consul-kv-sync -dir kv-files/production -consul-addr http://consul:8500 -verbose
```

When ACLs are enabled, the token is read from the `CONSUL_HTTP_TOKEN` environment variable, following the Consul CLI convention:

```bash
$ CONSUL_HTTP_TOKEN=<token> consul-kv-sync -dir kv-files/production
```

Export to JSON format:

```bash
$ consul-kv-sync -dir kv-files/production -export > production-kv.json
```

## Configuration

`consul-kv-sync` syncs the YAML files in whatever directory `-dir` points at. Organizing configuration by environment is a matter of directory naming, not a tool feature: keep one directory per environment and point `-dir` at the one you want to sync.

See the `example/` directory for a sample layout.

## How it Works

1. Loads every `.yaml` / `.yml` file directly inside `-dir`
2. Detects duplicate keys across files
3. Converts nested YAML structure to flat key-value pairs
4. Synchronizes to Consul using Transaction API in batches

The directory is the source of truth: to add configuration, drop a YAML file into it. There is no separate manifest to keep in sync.

## License

This project is licensed under the [MIT License](./LICENSE).
