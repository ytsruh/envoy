# Envoy CLI Usage Guide

## Overview

The Envoy CLI supports two distinct modes of operation:

- **Argument Mode**: Provide all required arguments for fast, scriptable operations
- **Interactive Mode**: Run commands without arguments for guided prompts

## Modes

### Argument Mode

Best for:
- Scripting and automation
- Quick operations when you know the IDs
- CI/CD pipelines
- Experienced users

**Usage:** Provide all required arguments

```bash
# Project commands
envoy projects get <project_id>
envoy projects update <project_id>
envoy projects delete <project_id>

# Environment commands
envoy environments create <project_id>
envoy environments list <project_id>
envoy environments get <environment_id> <project_id>
envoy environments update <environment_id> <project_id>
envoy environments delete <environment_id> <project_id>

# Variable commands
envoy variables create <project_id> <environment_id>
envoy variables list <project_id> <environment_id>
envoy variables get <variable_id> <project_id> <environment_id>
envoy variables update <variable_id> <project_id> <environment_id>
envoy variables delete <variable_id> <project_id> <environment_id>
envoy variables import -f .env
envoy variables export -f .env
```

**Examples:**

```bash
# List all environments for a project
envoy environments list 123e4567-e89b-12d3-a456-426614174000

# Get a specific variable's details
envoy variables get var-123 123e4567-e89b-12d3-a456-426614174000 env-456

# Update a project
envoy projects update 123e4567-e89b-12d3-a456-426614174000
```

### Interactive Mode

Best for:
- New users learning the CLI
- Exploring available resources
- Quick tasks without looking up IDs
- Manual operations

**Usage:** Run commands without any arguments

```bash
# Project commands
envoy projects get
envoy projects update
envoy projects delete

# Environment commands
envoy environments create
envoy environments list
envoy environments get
envoy environments update
envoy environments delete

# Variable commands
envoy variables create
envoy variables list
envoy variables get
envoy variables update
envoy variables delete
envoy variables import
envoy variables export
```

**Examples:**

```bash
# Create an environment (will prompt for project selection)
$ envoy environments create

Select a project:
  1. My Project - A sample project
  2. Another Project
  0. Cancel

Select an option [1-2]: 1
Environment name: production
Description (optional): Production environment

Environment created successfully!
  ID: env-123
  Name: production
  Project ID: proj-456
```

## Command-by-Command Examples

### Projects

```bash
# Argument mode
envoy projects list
envoy projects get 123e4567-e89b-12d3-a456-426614174000
envoy projects create  # Already interactive
envoy projects update 123e4567-e89b-12d3-a456-426614174000
envoy projects delete 123e4567-e89b-12d3-a456-426614174000

# Interactive mode
envoy projects get  # Prompts to select project
envoy projects update  # Prompts to select project, then update fields
envoy projects delete  # Prompts to select project, then confirms deletion
```

### Environments

```bash
# Argument mode
envoy environments create 123e4567-e89b-12d3-a456-426614174000
envoy environments list 123e4567-e89b-12d3-a456-426614174000
envoy environments get env-123 123e4567-e89b-12d3-a456-426614174000
envoy environments update env-123 123e4567-e89b-12d3-a456-426614174000
envoy environments delete env-123 123e4567-e89b-12d3-a456-426614174000

# Interactive mode
envoy environments create  # Prompts for project, then name/description
envoy environments list  # Prompts for project, then lists environments
envoy environments get  # Prompts for project, then environment, shows details
envoy environments update  # Prompts for project, environment, then updates
envoy environments delete  # Prompts for project, environment, confirms, deletes
```

### Variables

```bash
# Argument mode
envoy variables create 123e4567-e89b-12d3-a456-426614174000 env-123
envoy variables list 123e4567-e89b-12d3-a456-426614174000 env-123
envoy variables get var-456 123e4567-e89b-12d3-a456-426614174000 env-123
envoy variables update var-456 123e4567-e89b-12d3-a456-426614174000 env-123
envoy variables delete var-456 123e4567-e89b-12d3-a456-426614174000 env-123
envoy variables import -f .env
envoy variables export -f .env

# Interactive mode
envoy variables create  # Prompts for project, environment, key, value
envoy variables list  # Prompts for project, environment, lists variables
envoy variables get  # Prompts for project, environment, variable, shows details
envoy variables update  # Prompts for project, environment, variable, updates
envoy variables delete  # Prompts for project, environment, variable, confirms, deletes
envoy variables import  # Prompts for project, environment
envoy variables export  # Prompts for project, environment
```

## Important Notes

### No Mixed Mode

If you provide some but not all arguments, the CLI will show an error:

```bash
# This will fail
$ envoy environments get env-123

Error: Both environment_id and project_id are required
Usage: envoy environments get <environment_id> <project_id>
```

Use either all arguments (argument mode) or no arguments (interactive mode).

### Error Handling

**Argument Mode:**
- Invalid IDs show an error and exit
- No prompts for correction
- Best for scripts where you want clear failures

**Interactive Mode:**
- Invalid selections can be cancelled and retried
- Prompts guide you through the process
- Option 0 is always available to cancel

### Cancellation

All interactive prompts include a "Cancel" option (0):

```bash
Select a project:
  1. My Project
  2. Another Project
  0. Cancel

Select an option [1-2]: 0
Import cancelled
```

### Default Values

Update commands in both modes show current values as defaults:

```bash
$ envoy variables update

# After selecting project, environment, and variable...

Variable key [API_KEY]:
Variable value: secret123
```

Press Enter to keep the current value, or type a new value to change it.

## Tips

1. **Use `list` commands first** to find IDs you need for argument mode
2. **Interactive mode is great for exploration** - try it when learning the CLI
3. **Argument mode for scripts** - provide all IDs for reliable automation
4. **Copy IDs from list output** - they're displayed for easy copying
5. **Use tab completion** - if your shell supports it, IDs can be tab-completed (not yet implemented)
