# Transfer App - Rubix Blockchain Token Transfer Tool

A standalone Go application that directly communicates with the Rubix blockchain to perform token transfers. This application implements all cryptographic operations locally, including ECDSA signing and image-based signature generation, eliminating the need for the fexr-flutter SDK.

## Features

- **Direct Rubix Blockchain Integration**: Communicates directly with Rubix HTTP REST APIs
- **Local Cryptographic Operations**:
  - ECDSA signing (P-256 curve, ASN.1 DER encoding)
  - Image-based signature generation from PNG files
  - SHA3-256 hashing
- **Two-Phase Token Transfer**: Handles the complete transfer flow
- **File-Based Account Management**: Export DIDs with balances to JSON file and use for transfers
- **Automatic Balance Validation**: Prevents transfers with insufficient balance
- **CLI Interface**: Easy-to-use command-line interface
- **Key Management**: Generate and manage EC key pairs

## Architecture

```
break-nlss/
├── main.go                 # CLI entry point
├── pkg/
│   ├── crypto/            # Cryptographic operations
│   │   ├── hash.go       # SHA3-256 hashing
│   │   ├── ecdsa.go      # ECDSA key operations and signing
│   │   └── image.go      # Image-based signature generation
│   ├── rubix/            # Rubix blockchain client
│   │   ├── client.go     # HTTP client wrapper
│   │   ├── transaction.go # Token transfer operations
│   │   └── models.go     # Request/Response structs
│   ├── config/           # Configuration management
│   │   └── config.go
│   └── storage/          # File-based account storage
│       └── accounts.go   # DID account persistence
├── internal/
│   └── files/            # File loading utilities
│       └── loader.go
├── preset/               # User cryptographic files (you provide these)
│   ├── privatekey.pem
│   ├── PrivateShare.png
│   ├── PublicShare.png   # Optional
│   └── DID.png          # Optional
├── test/                 # Unit tests
│   └── crypto_test.go
└── accounts.json         # Exported DID accounts (generated)
```

## Quick Start

### Using File-Based Transfers (Recommended)

This is the easiest way to get started:

```bash
# 1. Export DIDs with balance > 0
./break-nlss export-dids --output accounts.json

# 2. View the exported accounts
cat accounts.json

# 3. Edit accounts.json and add your peer_id
nano accounts.json  # Add "peer_id": "12D3Koo..." for your sender DID

# 4. Copy your PrivateShare.png to preset folder
cp ~/.rubix/<SENDER_DID>/PrivateShare.png preset/

# 5. Execute transfer
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 5.0 \
  --comment "My first transfer"
```

### Using Environment Variables

Traditional approach with environment configuration:

```bash
# 1. Set environment variables
export RUBIX_NODE_URL="localhost:20006"
export SENDER_PEER_ID="12D3Koo..."
export SENDER_DID="bafybmi..."

# 2. Copy PrivateShare.png
cp ~/.rubix/$SENDER_DID/PrivateShare.png preset/

# 3. Execute transfer
./break-nlss transfer \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 5.0
```

## Installation

### Prerequisites

- Go 1.21 or higher
- Access to a Rubix blockchain node
- Your cryptographic files (private key and PrivateShare.png)

### Build from Source

```bash
# Clone the repository
cd /Users/allen/Professional/break-nlss

# Install dependencies
go mod download

# Build the application
go build -o break-nlss

# Or run directly
go run main.go
```

## Setup

### 1. Prepare the Preset Folder

The `preset/` folder must contain your cryptographic files:

**Required files:**
- `privatekey.pem` - Your EC private key (P-256 curve)
- `PrivateShare.png` - Your private image share for signature generation

**Optional files:**
- `PublicShare.png` - Your public image share
- `DID.png` - DID image representation

### 2. Set Environment Variables

```bash
export RUBIX_NODE_URL="localhost:20006"
export SENDER_PEER_ID="your_peer_id"
export SENDER_DID="your_did"
export PRESET_FOLDER="./preset"  # Optional, defaults to ./preset
```

### 3. Generate Keys (Optional)

If you don't have a private key yet, you can generate one:

```bash
./break-nlss generate-key --output ./preset
```

This will create:
- `preset/privatekey.pem` - Your new private key
- `preset/publickey.pem` - Corresponding public key

**⚠️ IMPORTANT: Keep your private key secure and never share it!**

## Usage

### Transfer Tokens

Transfer RBT tokens to another DID:

```bash
./break-nlss transfer \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.5 \
  --comment "Payment for services"
```

With explicit configuration:

```bash
./break-nlss transfer \
  --rubix-node localhost:20006 \
  --sender-peer peer789 \
  --sender-did bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.5 \
  --comment "Payment" \
  --preset ./preset
```

### Check Balance

Get the balance for a DID:

```bash
# Check your own balance (uses SENDER_DID from env)
./break-nlss balance

# Check another DID's balance
./break-nlss balance --did bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u

# Specify Rubix node
./break-nlss balance --rubix-node localhost:20006 --did bafybmi...
```

### List All DIDs

List all DIDs registered on the Rubix node:

```bash
./break-nlss list-dids

# With specific node
./break-nlss list-dids --rubix-node localhost:20006
```

### Export DIDs to File

Export DIDs with balance > 0 to a JSON file for file-based transfers:

```bash
# Export all DIDs with balance > 0
./break-nlss export-dids --output accounts.json

# Export with custom minimum balance filter
./break-nlss export-dids --output accounts.json --min-balance 10.0

# Export from specific node
./break-nlss export-dids --output my-accounts.json --rubix-node localhost:8080
```

**Output:** Creates `accounts.json` containing:
```json
{
  "version": "1.0",
  "rubix_node_url": "localhost:20006",
  "exported_at": "2025-11-19T18:53:39.111233+05:30",
  "total_dids": 3,
  "accounts": [
    {
      "did": "bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy",
      "peer_id": "",
      "balance": 67,
      "did_type": 4,
      "pledged_rbt": 0,
      "locked_rbt": 0.9,
      "pinned_rbt": 0,
      "updated_at": "2025-11-19T18:53:39.111228+05:30"
    }
  ]
}
```

**Important:** Manually add the `peer_id` for DIDs you want to use as senders.

### Transfer from File (File-Based Workflow)

Transfer tokens using sender information from exported accounts file:

```bash
# Transfer using account from file
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 5.0 \
  --comment "Payment from file"
```

**Benefits:**
- ✅ **Automatic Balance Check**: Validates sender has sufficient balance before transfer
- ✅ **No Environment Variables**: All sender info read from file
- ✅ **Multiple Senders**: Use different `--sender-index` for different accounts
- ✅ **Rubix Node Auto-Config**: Uses node URL from the file

**Complete File-Based Workflow:**

```bash
# Step 1: Export DIDs with balance > 0
./break-nlss export-dids --output accounts.json

# Step 2: Edit accounts.json and add peer_id for sender(s)
nano accounts.json  # Add "peer_id": "12D3Koo..." for the DID you'll use

# Step 3: Copy PrivateShare.png for the sender DID
cp ~/.rubix/<SENDER_DID>/PrivateShare.png preset/

# Step 4: Transfer using file
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmi... \
  --amount 10.0
```

### Generate Key Pair

Generate a new EC key pair (P-256 curve):

```bash
./break-nlss generate-key --output ./preset
```

## Commands Reference

Quick reference for all available commands:

| Command | Description | Key Options |
|---------|-------------|-------------|
| `export-dids` | Export DIDs with balance > 0 to JSON file | `--output`, `--min-balance`, `--rubix-node` |
| `transfer` | Transfer RBT tokens (two modes) | **Standard**: `--receiver`, `--amount`, `--sender-peer`, `--sender-did`<br>**File-based**: `--from-file`, `--sender-index`, `--receiver`, `--amount` |
| `balance` | Get account balance for a DID | `--did`, `--rubix-node` |
| `list-dids` | List all DIDs from the node | `--rubix-node` |
| `generate-key` | Generate new EC key pair | `--output` |
| `help` | Show help message | - |

### Transfer Modes Comparison

| Aspect | Standard Mode | File-Based Mode |
|--------|---------------|-----------------|
| **Sender Config** | Environment vars or flags | Read from JSON file |
| **Balance Check** | Manual | Automatic |
| **Multiple Senders** | Change env vars | Use different `--sender-index` |
| **Node Config** | Environment or flag | Auto from file |
| **Use Case** | Single sender, frequent use | Multiple senders, batch operations |

**Example Standard:**
```bash
export SENDER_PEER_ID="..." SENDER_DID="..."
./break-nlss transfer --receiver bafybmi... --amount 10
```

**Example File-Based:**
```bash
./break-nlss export-dids --output accounts.json
./break-nlss transfer --from-file accounts.json --sender-index 0 --receiver bafybmi... --amount 10
```

## How It Works

### Two-Phase Token Transfer

The application performs a two-phase token transfer:

#### Phase 1: Initiate Transfer
1. Call `POST /api/initiate-rbt-transfer` with transfer details
2. Receive transaction ID and hash to sign

#### Phase 2: Sign and Complete
1. Decode the Base64-encoded hash
2. Generate image-based signature from `PrivateShare.png`
3. Hash the image signature with SHA3-256
4. Sign the hash with ECDSA private key
5. Submit both signatures via `POST /api/signature-response`

### Cryptographic Operations

#### Image-Based Signature Generation

This is the most critical algorithm. It must match exactly with the Dart implementation:

1. **Convert PNG to Binary**: Extract RGB values from each pixel and convert to binary string
2. **Generate Positions**: Use deterministic formula to calculate 256 bit positions from hash:
   ```
   position = (((2402 + hashChar) * 2709) + ((k + 2709) + hashChar)) % 2048
   ```
3. **Extract Signature Bits**: Get bits at calculated positions from the image binary
4. **Convert to Bytes**: Pack the 256 bits into 32 bytes

#### ECDSA Signing

1. Load EC private key (P-256 curve) from PEM file
2. Hash the image signature with SHA3-256
3. Sign the hash using ECDSA
4. Encode signature as ASN.1 DER (SEQUENCE of R and S integers)

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RUBIX_NODE_URL` | Rubix blockchain node URL | `localhost:20006` |
| `SENDER_PEER_ID` | Your peer ID | (required) |
| `SENDER_DID` | Your DID | (required) |
| `PRESET_FOLDER` | Path to preset folder | `./preset` |

### Command-Line Flags

All environment variables can be overridden with command-line flags:

```bash
--rubix-node    Rubix node URL
--sender-peer   Sender peer ID
--sender-did    Sender DID
--preset        Preset folder path
```

## API Reference

### Rubix Blockchain APIs

#### Initiate Transfer
- **Endpoint**: `POST /api/initiate-rbt-transfer`
- **Request**:
  ```json
  {
    "receiver": "bafybmi...",
    "sender": "peer123.bafybmi...",
    "tokenCOunt": 10.5,
    "comment": "Payment",
    "type": 2
  }
  ```
  ⚠️ **Note**: Field name is `tokenCOunt` with capital 'O'

- **Response**:
  ```json
  {
    "status": true,
    "result": {
      "id": "txn_abc123",
      "hash": "SGVsbG8gV29ybGQh..."
    },
    "message": "Transaction initiated"
  }
  ```

#### Submit Signature
- **Endpoint**: `POST /api/signature-response`
- **Request**:
  ```json
  {
    "id": "txn_abc123",
    "signature": {
      "Signature": [48, 69, 2, 33, ...],
      "Pixels": [1, 0, 1, 0, ...]
    }
  }
  ```

- **Response**:
  ```json
  {
    "status": true,
    "message": "Transaction completed successfully"
  }
  ```

#### Get Account Balance
- **Endpoint**: `GET /api/get-account-info?did={did}`
- **Response**:
  ```json
  {
    "status": true,
    "account_info": [{
      "did": "bafybmi...",
      "did_type": 4,
      "rbt_amount": 67.0,
      "pledged_rbt": 0,
      "locked_rbt": 0.9,
      "pinned_rbt": 0
    }]
  }
  ```

## Testing

### Run Unit Tests

```bash
# Run all tests
go test ./test/...

# Run with verbose output
go test -v ./test/...

# Run specific test
go test -v ./test/ -run TestCalculateSHA3Hash
```

### Test Coverage

```bash
go test -cover ./pkg/...
```

## Troubleshooting

### Common Issues

1. **"Configuration error: SENDER_PEER_ID is required"**
   - Set the `SENDER_PEER_ID` environment variable or use `--sender-peer` flag
   - **OR** use file-based transfer: `--from-file accounts.json --sender-index 0`

2. **"Private Key file does not exist"**
   - Ensure `privatekey.pem` exists in your preset folder
   - Generate a new key with `break-nlss generate-key`

3. **"PrivateShare.png file does not exist"**
   - You must provide your own `PrivateShare.png` file
   - This file is unique to your DID and cannot be generated
   - Copy from: `~/.rubix/<YOUR_DID>/PrivateShare.png`

4. **"failed to initiate transfer"**
   - Check that Rubix node is running and accessible
   - Verify the node URL is correct
   - Ensure you have sufficient balance

5. **"signature submission failed"**
   - Verify your private key matches your DID
   - Ensure PrivateShare.png is correct for your DID

6. **"Error: Insufficient balance"** (File-based transfer)
   - The sender account doesn't have enough RBT
   - Check balance with: `./break-nlss balance --did <SENDER_DID>`
   - Update accounts file: `./break-nlss export-dids --output accounts.json`

7. **"Invalid sender index"** (File-based transfer)
   - The index specified doesn't exist in the accounts file
   - Check accounts file: `cat accounts.json | jq '.accounts | length'`
   - Remember: indexes are 0-based (first account is index 0)

8. **"Error loading accounts file"**
   - Ensure the accounts file exists and is valid JSON
   - Re-export: `./break-nlss export-dids --output accounts.json`
   - Verify JSON format: `cat accounts.json | jq .`

### Debug Mode

Enable verbose logging:

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o break-nlss

# Run with verbose output
./break-nlss transfer --receiver ... --amount ... -v
```

## Development

### Project Structure

- `pkg/crypto/` - Core cryptographic primitives
- `pkg/rubix/` - Rubix blockchain client and transaction handling
- `pkg/config/` - Configuration management
- `pkg/storage/` - File-based account storage and management
- `internal/files/` - Internal file utilities
- `test/` - Unit tests

### Adding New Features

1. Add new functionality to appropriate package
2. Write unit tests
3. Update README documentation
4. Test with real Rubix node

## Security Considerations

1. **Private Key Security**
   - Store private keys securely with restricted permissions (chmod 600)
   - Never commit private keys to version control
   - Use environment variables for sensitive configuration

2. **Image Shares**
   - PrivateShare.png must be kept secret
   - PublicShare.png can be shared safely

3. **Network Security**
   - Use HTTPS when connecting to remote Rubix nodes
   - Validate SSL certificates in production

## References

- **Sky Server**: `/Users/allen/Professional/sky`
- **fexr-flutter SDK**: `/Users/allen/Professional/sky/fexr-flutter`
- **Dart Implementations**:
  - Image processing: `lib/signature/dependencies.dart`
  - Signature generation: `lib/signature/gen_sign.dart`
  - ECDSA operations: `lib/signature/key_gen.dart`
  - Rubix APIs: `lib/native_interaction/rubix/rubix_platform_calls.dart`

## License

[Your License Here]

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## Support

For issues and questions:
- Check the Troubleshooting section
- Review the API Reference
- Check Rubix blockchain documentation

---

**Version**: 1.0.0
**Last Updated**: 2025-11-19
