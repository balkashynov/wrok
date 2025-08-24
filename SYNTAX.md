# Wrok Task Syntax Guide

Wrok supports intelligent parsing of task descriptions with embedded metadata using special syntax. This allows you to quickly create rich tasks without using separate flags.

## Smart Parsing Syntax

### Tags: `#tag`
- Use `#` followed by the tag name
- Multiple tags supported
- Examples: `#backend #urgent #bug`

### Project: `@project`
- Use `@` followed by the project name  
- Only one project per task
- Examples: `@frontend @auth-service @documentation`

### Priority: `+priority`
- Use `+` followed by priority level
- Supports: `+low`, `+medium`, `+high`, `+1`, `+2`, `+3`
- Priority mapping: `1`=low, `2`=medium, `3`=high
- Examples: `+high`, `+2`, `+medium`

### JIRA Ticket: `ABC-123`
- Automatically detects JIRA ticket patterns
- Format: 3+ letters, dash, 1+ numbers
- Case insensitive (stored as uppercase)
- Examples: `DEV-456`, `proj-789`, `ABC-1234`

## Command Examples

### Interactive Mode (TUI)
```bash
# Launch beautiful interactive interface
wrok add

# Pre-fill with smart parsing
wrok add "Fix login bug #backend @auth +high ABC-123"
```

### Non-Interactive Mode
```bash
# Basic smart parsing
wrok add "Fix login bug #backend #urgent @auth +high ABC-123"

# With additional flags
wrok add "Update schema #db +medium DEV-456" --due "3 days" --note "Include migration"

# Mix of smart parsing and flags
wrok add "Write docs @documentation #writing" --priority high --due "1 week"

# Simple task without parsing
wrok add "Simple task without metadata"
```

## Parsing Examples

| Input | Result |
|-------|--------|
| `Fix login bug #backend #urgent @auth !high ABC-123` | Title: "Fix login bug !high"<br>Project: auth<br>Tags: backend, urgent<br>JIRA: ABC-123<br>Priority: high |
| `Update database schema #db !medium DEV-456` | Title: "Update database schema !medium"<br>Tags: db<br>JIRA: DEV-456<br>Priority: medium |
| `Write docs @documentation !1 #writing #markdown` | Title: "Write docs !1"<br>Project: documentation<br>Tags: writing, markdown<br>Priority: low |
| `Simple task without any parsing` | Title: "Simple task without any parsing"<br>(no metadata) |

## Available Flags

When using non-interactive mode, you can combine smart parsing with explicit flags:

- `--due string` - Due date: dd/mm/yyyy, X days, X hours, X weeks
- `--jira string` - JIRA ticket ID
- `--note string` - Additional notes  
- `--priority string` - Priority: low, medium, high, or 1-3
- `-p, --project string` - Project name
- `-t, --tags strings` - Comma-separated tags
- `--url string` - Related URL
- `-i, --interactive` - Force interactive mode

## Due Date Formats

Due dates support multiple flexible formats:

- **Absolute**: `15/12/2024`, `31/01/2025`
- **Relative**: `3 days`, `1 week`, `24 hours`, `2 weeks`

Examples:
```bash
wrok add "Review PR #code-review" --due "tomorrow"
wrok add "Deploy to prod #deployment !high" --due "31/12/2024"  
wrok add "Weekly sync @team" --due "1 week"
```

## Pro Tips

1. **Order doesn't matter**: `#tag @project !priority` or `!priority #tag @project` both work
2. **Flexible spacing**: Spaces around symbols are optional
3. **Case insensitive JIRA**: `abc-123` becomes `ABC-123`
4. **Priority display**: Priority is kept in the title for visual reference
5. **Combine with flags**: Smart parsing + flags gives you maximum flexibility
6. **Interactive preview**: Use interactive mode to see real-time parsing as you type

The smart parsing makes task creation incredibly fast while maintaining full flexibility through explicit flags when needed.