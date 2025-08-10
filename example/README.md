# consul-kv-sync Example Configuration

This directory contains example configurations demonstrating how to use consul-kv-sync for different environments.

## Quick Test

From this directory, you can test the tool with:

```bash
# Dry run for development environment
$ consul-kv-sync -env development -dry-run

# Dry run for staging environment
$ consul-kv-sync -env staging -dry-run

# Dry run for production environment
$ consul-kv-sync -env production -dry-run
```

## Structure

```
example/
├── environments.yaml   # Environment definitions
└── kv-files/           # KV configurations per environment
    ├── development/    # Development configs (minimal)
    ├── staging/        # Staging configs (feature flags enabled)
    └── production/     # Production configs (full setup)
```
