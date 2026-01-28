# Locara - Archive Server

Simple archive server built in Go with embedded frontend.

A rewrite of an old version of this software that was written in Node.JS + Express + Vue.

## Features

- File upload with metadata (name, date, type, author)
- Archive browsing grouped by year
- Download archives
- Simple authorization code system
- Theme switching (default and ayu_mirage)
- Single binary with embedded assets
- Configuration via TOML

## Installation

### From source

```bash
git clone https://github.com/Firstbober/locara.git
cd locara
go build -o locara ./cmd/locara
```

## Configuration

Create a `config.toml` file:

```toml
use_directory = "./uploads"
port = 4000
base_url = "" # i.e. when running under /locara and not /

[[users]]
name = "username"
auth = "your_auth_code"
```

## Usage

### Development mode

```bash
# Run with default config
./locara

# Custom config file
./locara -config /path/to/config.toml

# Custom port
./locara -port 8080
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | / | Index page (list archives) |
| GET | /upload | Upload form |
| POST | /api/archive/create | Upload new archive |
| GET | /api/archives | JSON list of all archives |
| GET | /api/archive/{id} | Download archive file |

## File Storage

Archives are stored in the configured directory:

```
uploads/
├── 1/
│   ├── info.json (metadata)
│   └── filename.ext (actual file)
├── 2/
│   ├── info.json
│   └── filename.ext
└── ...
```

## Development

### Running tests

```bash
go test ./...
```

### Building and Running

```bash
bunx nodemon --watch './**/*' -e go,html --signal SIGTERM --exec 'go' run cmd/locara/main.go
```

## License

See LICENSE file.
