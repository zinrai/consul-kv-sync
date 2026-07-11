# consul-kv-sync Example Configuration

This directory contains example configurations demonstrating how to use consul-kv-sync for different environments.

## Quick Test

From this directory, you can test the tool with:

```bash
# Dry run for development configs
$ consul-kv-sync -dir kv-files/development -dry-run

# Dry run for staging configs
$ consul-kv-sync -dir kv-files/staging -dry-run

# Dry run for production configs
$ consul-kv-sync -dir kv-files/production -dry-run
```

## Structure

```
example/
└── kv-files/           # One directory per environment
    ├── development/    # Development configs (minimal)
    ├── staging/        # Staging configs (feature flags enabled)
    └── production/     # Production configs (full setup)
```

Point `-dir` at one of these directories to sync every `.yaml` / `.yml` file inside it. The directory names are just a convention: the tool syncs whatever directory it is given.
