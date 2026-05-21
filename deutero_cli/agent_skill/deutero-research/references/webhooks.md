# Webhooks Reference

Complete reference for managing Deutero webhook endpoints via `deutero-cli`.

---

## Overview

Webhooks allow external systems to receive real-time notifications when events occur in Deutero (e.g. an interview is completed). Each webhook endpoint is a URL that Deutero POSTs a signed event payload to when a subscribed event fires.

**Important:** Webhook endpoints must be **created** via the [Deutero dashboard](https://app.deutero.ai). The CLI can list, update, and delete existing endpoints — but not create new ones. Signing secrets are never exposed through the CLI; retrieve them from the dashboard.

---

## List Webhook Endpoints

```bash
deutero webhooks list
deutero webhooks list -o webhooks.json
```

Returns a table with columns: `id`, `label`, `url`, `enabled`, `events`.

Also prints the list of **available event types** that endpoints can subscribe to.

### Available Event Types

| Event | Fired when |
|-------|-----------|
| `interview.completed` | A participant finishes an interview (real or simulated) |
| `survey.completed` | All interviews in a survey reach completion or quota is filled |

Additional event types may be available — run `deutero webhooks list` to see the current list returned by the API.

---

## Update a Webhook Endpoint

```bash
deutero webhooks update <ENDPOINT_ID> [options]
```

At least one option must be provided. Only supplied fields are changed; omitted fields remain untouched.

### Options

| Option | Description |
|--------|-------------|
| `--label <text>` | New human-readable label for the endpoint |
| `--url <url>` | New delivery URL (must be `http://` or `https://`) |
| `--enable` | Enable the endpoint (it will receive events) |
| `--disable` | Disable the endpoint (events are not delivered) |
| `--event <type>` | Subscribe to an event type — **repeatable**, **replaces** the existing event list |
| `-o, --output <file>` | Write JSON response to a file |

### Examples

```bash
# Rename an endpoint
deutero webhooks update <ENDPOINT_ID> --label "Production Events"

# Change the delivery URL
deutero webhooks update <ENDPOINT_ID> --url https://hooks.example.com/deutero

# Enable a disabled endpoint
deutero webhooks update <ENDPOINT_ID> --enable

# Disable an endpoint (e.g. during maintenance)
deutero webhooks update <ENDPOINT_ID> --disable

# Subscribe to specific events (replaces the current event list)
deutero webhooks update <ENDPOINT_ID> \
  --event interview.completed \
  --event survey.completed

# Subscribe to a single event only
deutero webhooks update <ENDPOINT_ID> --event interview.completed

# Combine multiple updates in one call
deutero webhooks update <ENDPOINT_ID> \
  --label "Updated Hook" \
  --enable \
  --event interview.completed \
  -o result.json
```

### Notes on `--enable` / `--disable`

- `--enable` and `--disable` are separate `is_flag` options with distinct parameter names
- When **both** are passed, `--enable` takes effect
- When neither is passed, the `enabled` field is not changed

### Notes on `--event`

- The `--event` flag is repeatable: use it multiple times for multiple event types
- Providing any `--event` flags **replaces** the endpoint's entire event subscription list
- To unsubscribe from all events, update via the dashboard (the CLI requires at least one option, but an empty event list is not supported via CLI)

---

## Delete a Webhook Endpoint

```bash
deutero webhooks delete <ENDPOINT_ID>
deutero webhooks delete <ENDPOINT_ID> --yes
deutero webhooks delete <ENDPOINT_ID> -y
```

This permanently deletes the endpoint and all its delivery logs. The operation **cannot be undone**.

By default, the CLI prompts for confirmation:
```
Delete webhook endpoint <ENDPOINT_ID>? This cannot be undone. [y/N]:
```

Pass `--yes` / `-y` to skip the confirmation prompt (useful in scripts or CI).

---

## Typical Webhook Workflow

```bash
# 1. List endpoints and see available event types
deutero webhooks list

# 2. Update endpoint to subscribe to relevant events
deutero webhooks update <ENDPOINT_ID> \
  --event interview.completed \
  --event survey.completed \
  --enable

# 3. Verify the update
deutero webhooks list

# 4. During maintenance, temporarily disable
deutero webhooks update <ENDPOINT_ID> --disable

# 5. Re-enable after maintenance
deutero webhooks update <ENDPOINT_ID> --enable

# 6. Clean up a decommissioned endpoint
deutero webhooks delete <ENDPOINT_ID> --yes
```

---

## Webhook Payload Security

Signing secrets are provisioned via the Deutero dashboard and are never returned by the CLI. To verify incoming webhook payloads:

1. Log in to [app.deutero.ai](https://app.deutero.ai)
2. Navigate to **Webhooks** → select the endpoint
3. Copy the signing secret
4. Verify the HMAC signature on incoming POST requests using the secret

This ensures only genuine Deutero events are processed by your system.
