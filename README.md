# Backpack — Development & Deployment

## Prerequisites

- Go 1.26+
- SQLite
- [Caddy](https://caddyserver.com/docs/install#debian-ubuntu-raspbian)
- [Just](https://just.systems/man/en/installation.html) _Optional_
- A Hetzner VPS (or any Linux server)
- A domain with an A record pointing to the server

---

## Development

Clone the repository:

```bash
git clone https://github.com/Ayomided/jmcann-suffolk-backpack
cd jmcann-suffolk-backpack
```

Generate a JWT_SECRET and add to .env
```bash
just gen-jwt // openssl rand -base64 32
```

Run the database migration:
```bash
just migrate // go run ./cmd/backend/main.go -db_path=backpack-app.db -migrate
```

Seed the database with test data:

```bash
just seed // go run ./cmd/backend/main.go -db_path=backpack-app.db -seed
```

Start the development server:

```bash
just dev // go run ./cmd/backend/main.go -addr :3000 -db_path=backpack-app.db
```

The application will be available at `http://localhost:3000`.

Default credentials after seeding:

```
Email:    qs@backpack.dev
Password: password123
```

## Justfile Commands

| Command        | Description                        | Runs                                                              |
|----------------|------------------------------------|-------------------------------------------------------------------|
| `just gen-jwt` | Generate a JWT secrets             | `openssl rand -base64 32`                                         |
| `just migrate` | Run the database migration         | `go run ./cmd/backend/main.go -db_path=backpack-app.db -migrate`  |
| `just seed`    | Seed the database with test data   | `go run ./cmd/backend/main.go -db_path=backpack-app.db -seed`     |
| `just dev`     | Start the development server       | `go run ./cmd/backend/main.go -addr :3000 -db_path=backpack-app.db` |
| `just build`   | Build the binary                   | `go build -o backpack-app ./cmd/backend/main.go`                  |
| `just test`    | Run all tests                      | `go test ./...`                                                   |
| `just clean`   | Remove the binary and database     | `rm -f backpack-app backpack-app.db`                              |
| `just reset`   | Clean, migrate, seed and start     | `just clean && just migrate && just seed && just dev`             |

> [!NOTE]
> If you get `JWT_SECRET not set` ensure your `.env` file exists and contains a valid secret. Re-run `just gen-jwt` to generate one, or prefix the command manually:
> ```bash
> JWT_SECRET=your-secret-here go run …
> ```

## Deployment

### Server Setup

Create a dedicated user on the server:

```bash
useradd -m -s /bin/bash backpack
mkdir -p /opt/backpack
chown -R backpack:backpack /opt/backpack
```

Grant the user permission to manage the systemd service:

```bash
visudo -f /etc/sudoers.d/backpack
```

```
backpack ALL=(ALL) NOPASSWD: /bin/systemctl restart backpack
backpack ALL=(ALL) NOPASSWD: /bin/systemctl stop backpack
backpack ALL=(ALL) NOPASSWD: /bin/systemctl start backpack
backpack ALL=(ALL) NOPASSWD: /bin/systemctl enable backpack
backpack ALL=(ALL) NOPASSWD: /bin/systemctl daemon-reload
backpack ALL=(ALL) NOPASSWD: /bin/mv /tmp/backpack.service /etc/systemd/system/backpack.service
```

### Caddy

Install Caddy and configure the reverse proxy:

```bash
sudo apt install -y caddy
sudo nano /etc/caddy/Caddyfile
```

```
<domain-name-here> {
    reverse_proxy localhost:3000
}
```

```bash
sudo systemctl enable caddy
sudo systemctl start caddy
```

### First Deploy

Run the migration on the server before starting the service for the first time:

```bash
/opt/backpack/backpack-app -db_path=/opt/backpack/backpack.db -migrate
```

Then enable and start the service:

```bash
sudo systemctl enable backpack
sudo systemctl start backpack
```

### CI/CD Pipeline

The pipeline runs on every push to `main` via GitHub Actions. It performs the following steps:

1. Run tests
1. Build the Go binary for Linux
1. Rsync the binary, templates and static assets to the server over SSH
1. Run `deploy.sh` on the server which stops the service, sets permissions, starts the service and health checks the application

The following secrets must be set in the repository settings:

| Secret            | Description                        |
|-------------------|------------------------------------|
| `HETZNER_HOST`    | IP address or hostname of the VPS  |
| `HETZNER_USER`    | SSH user on the server             |
| `HETZNER_SSH_KEY` | Private key for SSH authentication |

### Environment Variables

The application reads the following environment variables at startup, set in the systemd service file:

| Variable        | Description                              |
|-----------------|------------------------------------------|
| `JWT_SECRET`    | Secret used to sign JWT tokens           |
| `SECURE_COOKIE` | Set to `true` in production              |

---

## Health Check

The deploy script polls `https://backpack.adediiji.uk/login` after restarting the service. If the application does not respond within 10 attempts the script exits with a non-zero status, failing the pipeline and alerting via GitHub Actions.
