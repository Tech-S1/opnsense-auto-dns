# OPNsense Auto DNS

A CLI tool that automatically updates DNS records in OPNsense unbound. This tool can run once or continuously to keep your DNS records updated with your current IP address.

## Installation

### Prerequisites

- OPNsense firewall with API access enabled
- API key and secret from OPNsense

### Method 1: Download from Releases (Recommended)

1. Go to the [Releases page](https://github.com/your-username/opnsense-auto-dns/releases)
2. Download the appropriate archive for your platform:
   - **Linux (AMD64)**: `opnsense-auto-dns-linux-amd64.tar.gz`
   - **Linux (ARM64)**: `opnsense-auto-dns-linux-arm64.tar.gz`
   - **Windows**: `opnsense-auto-dns-windows-amd64.tar.gz`
   - **macOS (Intel)**: `opnsense-auto-dns-darwin-amd64.tar.gz`
   - **macOS (Apple Silicon)**: `opnsense-auto-dns-darwin-arm64.tar.gz`

3. Extract the archive:
```bash
tar -xzf opnsense-auto-dns-linux-amd64.tar.gz
```

4. Make the binary executable (Linux/macOS):
```bash
chmod +x opnsense-auto-dns
```

5. Move to a directory in your PATH (optional):
```bash
sudo mv opnsense-auto-dns /usr/local/bin/
```

### Method 2: Using Configuration

1. Copy the example configuration:
```bash
cp config.example.json config.json
```

2. Edit the configuration file with your settings:
```bash
nano config.json
```

3. Run the tool:
```bash
./opnsense-auto-dns auto-updater --config config.json
```

### Method 3: Build from Source

#### Prerequisites for Building

- Go 1.24.1 or later

#### Building

```bash
# Clone the repository
git clone <repository-url>
cd opnsense-auto-dns

# Build the binary
make build

# Or build manually
go build -o bin/opnsense-auto-dns .
```

#### Using Makefile

```bash
# Build the binary
make build

# Run the tool
make run

# Clean build artifacts
make clean

# Tidy Go modules
make tidy
```

## Configuration

### Configuration Precedence

The tool uses the following precedence order (highest to lowest):
1. Environment variables
2. Command line flags
3. Config file (if specified)

### Configuration Methods

#### 1. Config File

The easiest way to configure the tool is using a JSON configuration file. Use the provided example as a starting point:

```json
{
  "opnsense_host": "192.168.1.1",
  "opnsense_api_key": "your-api-key-here",
  "opnsense_api_secret": "your-api-secret-here",
  "domain": "example.com",
  "hostnames": ["server1", "server2"],
  "ip_address": "203.0.113.1"
}
```

#### 2. Environment Variables

Perfect for containerized deployments or when you prefer environment-based configuration:

```bash
export OPNSENSE_HOST="192.168.1.1"
export OPNSENSE_API_KEY="your-api-key-here"
export OPNSENSE_API_SECRET="your-api-secret-here"
export DOMAIN="example.com"
export HOSTNAMES="server1,server2"
export IP_ADDRESS="203.0.113.1"
export INTERVAL="5"
export LOOP="true"
export IGNORE_CERT="false"
```

#### 3. Command Line Flags

Useful for overriding specific settings or when running ad-hoc commands:

```bash
./opnsense-auto-dns auto-updater \
  --config config.json \
  --opnsense-host 192.168.1.1 \
  --opnsense-api-key your-api-key \
  --opnsense-api-secret your-api-secret \
  --domain example.com \
  --hostnames server1,server2 \
  --ip-address 203.0.113.1 \
  --interval 5 \
  --loop \
  --ignore-cert
```

### Required Configuration

The following parameters are required:

- **OPNsense Host**: IP address or hostname of your OPNsense firewall
- **API Key**: OPNsense API key
- **API Secret**: OPNsense API secret
- **Domain**: Domain name for DNS records

### Optional Configuration

- **Hostnames**: List of hostnames to update (defaults to machine hostname if not specified)
- **IP Address**: Specific IP address to use for DNS records (defaults to auto-detected machine IP if not specified)
- **Interval**: Update interval in minutes when running in loop mode (default: 5)
- **Loop**: Run continuously (default: false)
- **Ignore Cert**: Ignore SSL certificate validation (default: false)

### IP Address Configuration

The tool can use either a manually specified IP address or automatically detect the current machine's IP address:

- **Auto-detection (default)**: The tool automatically detects your machine's public IP address by creating a UDP connection to `1.1.1.1:80`
- **Manual specification**: You can specify a custom IP address using:
  - Config file: `"ip_address": "203.0.113.1"`
  - Environment variable: `IP_ADDRESS="203.0.113.1"`
  - Command line flag: `--ip-address 203.0.113.1`

This is useful when:
- You want to use a different IP address than the machine's detected IP
- You're running the tool on a machine behind NAT and want to use the public IP
- You want to point DNS records to a specific IP address




## OPNsense Setup

### 1. Enable API Access

1. Log into your OPNsense web interface
2. Navigate to **System > Access > Users**
3. Create a new user or use an existing one
4. Navigate to **System > Access > API**
5. Enable the API and configure access for your user

### 2. Generate API Credentials

1. In the API settings, generate an API key and secret for your user
2. Note down both the key and secret - you'll need them for configuration

### 3. Configure Unbound DNS

1. Navigate to **Services > Unbound DNS > Overrides**
2. Ensure you have the necessary permissions to create/modify DNS records
3. The tool will automatically create or update host overrides as needed

## Examples

### Example 1: Simple Single Run

```bash
# Create a simple config file
cat > config.json << EOF
{
  "opnsense_host": "192.168.1.1",
  "opnsense_api_key": "your-api-key",
  "opnsense_api_secret": "your-api-secret",
  "domain": "home.local",
  "ip_address": "203.0.113.1"
}
EOF

# Run the tool
./opnsense-auto-dns auto-updater --config config.json
```

### Example 2: Multiple Hostnames

```bash
# Config with multiple hostnames
cat > config.json << EOF
{
  "opnsense_host": "192.168.1.1",
  "opnsense_api_key": "your-api-key",
  "opnsense_api_secret": "your-api-secret",
  "domain": "home.local",
  "hostnames": ["server", "nas", "router"]
}
EOF

# Run in loop mode
./opnsense-auto-dns auto-updater --config config.json --loop --interval 10
```

### Example 3: Environment Variables Only

```bash
# Set environment variables
export OPNSENSE_HOST="192.168.1.1"
export OPNSENSE_API_KEY="your-api-key"
export OPNSENSE_API_SECRET="your-api-secret"
export DOMAIN="home.local"
export HOSTNAMES="server,nas"
export IP_ADDRESS="203.0.113.1"
export LOOP="true"
export INTERVAL="5"

# Run without config file
./opnsense-auto-dns auto-updater
```

## Logging

The tool provides comprehensive logging with configurable levels:

- **debug**: Detailed debug information
- **info**: General information (default)
- **warn**: Warning messages
- **error**: Error messages only

```bash
# Set log level
./opnsense-auto-dns --log-level debug auto-updater --config config.json
```


### Creating Releases

This project uses GitHub Actions to automatically build and release binaries for multiple platforms. To create a release:

1. Create and push a new tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. The GitHub Action will automatically:
   - Build binaries for Linux (AMD64/ARM64), Windows (AMD64), and macOS (AMD64/ARM64)
   - Create a GitHub release with all binaries
   - Generate release notes

The release will include:
- Compressed archives with binaries and documentation
- Individual binary files for each platform
- Automatic release notes based on commits
