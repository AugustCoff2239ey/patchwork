# patchwork

Lightweight diff-based config migration tool that tracks infrastructure configuration drift over time.

---

## Installation

```bash
go install github.com/yourorg/patchwork@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/patchwork.git && cd patchwork && go build ./...
```

---

## Usage

Initialize patchwork in your config directory, then snapshot and compare configurations over time.

```bash
# Initialize a new patchwork project
patchwork init --config ./infra/config

# Take a snapshot of the current configuration state
patchwork snapshot --name "pre-deploy-v2.1"

# Diff current state against a previous snapshot
patchwork diff --from "pre-deploy-v2.1" --to latest

# Apply a migration patch to reconcile drift
patchwork apply --patch ./patches/fix-network-policy.patch
```

Example output:

```
[~] services/api.yaml
  - replicas: 2
  + replicas: 4

[+] services/cache.yaml  (new)
[!] services/db.yaml     (drift detected)
```

---

## Why patchwork?

Infrastructure configs change constantly. Patchwork gives you a lightweight, Git-friendly audit trail of what changed, when, and how — without requiring a full orchestration platform.

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss significant changes.

---

## License

[MIT](LICENSE)