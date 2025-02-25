# move-forward

Move Forward is a CLI (or should we call it a starter kit?) for you let your customers run your webhooks locally easily. Built on top of [AbacatePay](https://github.com/AbacatePay/abacatepay-cli), Move Forward is a simple way to listen webhooks locally and test them without the need to deploy their webhook to the public internet.

## Features

- 🔒 Secure webhook testing without public internet exposure
- 🔄 Real-time webhook forwarding to your local server
- 🛠️ Support for multiple webhook providers
- 📱 Easy device-based authentication
- 💾 Persistent configuration management

## Installation

### macOS

##### For Intel Macs

```bash
curl -L https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-darwin-amd64 -o /usr/local/bin/move-forward
chmod +x /usr/local/bin/move-forward
```

##### For Apple Silicon Macs

```bash
curl -L https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-darwin-arm64 -o /usr/local/bin/move-forward
chmod +x /usr/local/bin/move-forward
```

### Linux

##### For AMD64

```bash
curl -L https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-linux-amd64 -o /usr/local/bin/move-forward
chmod +x /usr/local/bin/move-forward
```

##### For ARM64

```bash
curl -L https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-linux-arm64 -o /usr/local/bin/move-forward
chmod +x /usr/local/bin/move-forward
```

### Windows

1. Download the appropriate binary for your system:
   - [Windows AMD64](https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-windows-amd64.exe)
   - [Windows ARM64](https://github.com/brnocorreia/move-forward/releases/latest/download/move-forward-windows-arm64.exe)
2. Rename the file to `move-forward.exe`
3. Add it to your system's PATH

## Quick Start

1. List available webhook providers:

   ```bash
   move-forward list
   ```

2. Set up your preferred webhook provider:

   ```bash
   move-forward setup <provider-name>
   ```

3. Authenticate with the provider:

   ```bash
   move-forward login
   ```

4. Start listening for webhooks:
   ```bash
   move-forward listen --forward-url http://localhost:3000/webhook
   ```

## Usage

### Available Commands

- `move-forward list` - Display all available webhook providers
- `move-forward setup <provider>` - Configure the CLI for a specific webhook provider
- `move-forward login` - Authenticate with the configured provider
- `move-forward listen` - Start listening for webhooks

### Listen Command Options

```bash
move-forward listen [flags]
```

#### Flags

- `--forward-url` - URL to forward webhooks to (default: last used URL)

## Configuration

Move Forward stores its configuration in `~/.move-forward.json`. This includes:

- Authentication tokens
- Current webhook provider
- Forward URL
- Provider-specific settings

## Supported Webhook Providers

Move Forward supports various webhook providers out of the box. Use `move-forward list` to see all available providers.

## Development

### Prerequisites

- Go 1.21 or higher
- Git

### Building from Source

```bash
git clone https://github.com/brnocorreia/move-forward.git
cd cli
go build -o move-forward main.go
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
