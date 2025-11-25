# Break-NLSS - Rubix Blockchain Token Transfer Tool

A standalone Go application for Rubix blockchain token transfers with integrated NLSS (Non-Linear Secret Sharing) support. This tool reconstructs private shares from DIDs and performs cryptographically signed token transfers without requiring the fexr-flutter SDK.

## Quick Start

### Option 1: Simple Direct Transfer (Single Sender)

```bash
# 1. Build the application
go build -o break-nlss

# 2. Configure environment (.env file)
RUBIX_NODE_URL=localhost:20006
SENDER_DID=bafybmi...
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1

# 3. Reconstruct private share for your DID
./break-nlss break-nlss --did bafybmi...

# 4. Transfer tokens directly
./break-nlss transfer \
  --receiver bafybmi... \
  --amount 10.0 \
  --comment "Payment"
```

### Option 2: File-Based Transfer (Multiple Senders)

```bash
# 1. Build the application
go build -o break-nlss

# 2. Configure environment (.env file)
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1
RUBIX_NODE_URL=localhost:20006

# 3. Reconstruct private shares for your DIDs
./break-nlss break-nlss --did dids.txt

# 4. Export DIDs with balances
./break-nlss export-dids --output accounts.json

# 5. Transfer tokens using file
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmi... \
  --amount 10.0
```

---

## Commands Reference

### Overview

| Command | Description |
|---------|-------------|
| [`break-nlss`](#1-break-nlss) | Reconstruct private share from DID and public share |
| [`transfer`](#2-transfer) | Transfer RBT tokens to another DID |
| [`balance`](#3-balance) | Get account balance for a DID |
| [`list-dids`](#4-list-dids) | List all DIDs from the Rubix node |
| [`export-dids`](#5-export-dids) | Export DIDs with balance > 0 to JSON file |
| [`generate-key`](#6-generate-key) | Generate new EC key pair (P-256) |
| [`help`](#7-help) | Show help message |

---

### 1. break-nlss

Reconstructs private share images from DID and public share using the Break-NLSS algorithm.

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--did` | string | ✓ | Single DID string OR path to file containing DIDs (one per line) |

#### Configuration (from .env)

- `NLSS_BASE_PATH` - Base path for Rubix data (e.g., `/Users/allen/Professional/sky`)
- `NLSS_NODE_NAME` - Node name (e.g., `node1`)
- `NLSS_OUTPUT_DIR` - Output directory for generated pvtShare.png files (default: `./output`)

#### Input Path Structure

The tool expects the following directory structure for each DID:
```
{NLSS_BASE_PATH}/{NLSS_NODE_NAME}/Rubix/{DID}/
├── DIDImg.png       # DID image
└── pubShare.png     # Public share image
```

#### Output Path Structure

Generated private shares are saved to:
```
{NLSS_OUTPUT_DIR}/{DID}/pvtShare.png
```

#### Examples

**Single DID:**
```bash
./break-nlss break-nlss --did bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse
```

**Multiple DIDs from file:**
```bash
# Create dids.txt with one DID per line
cat > dids.txt <<EOF
bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse
bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy
bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
EOF

./break-nlss break-nlss --did dids.txt
```

**Sample dids.txt format:**
```
# Comments are supported (lines starting with #)
bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse
bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy

# Empty lines are ignored
bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
```

#### Output

```
[1/3] Processing DID: bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse
============================================
  DID Image: /Users/allen/Professional/sky/node1/Rubix/bafybmi.../DIDImg.png
  Public Share: /Users/allen/Professional/sky/node1/Rubix/bafybmi.../pubShare.png
  Output: ./output/bafybmi.../pvtShare.png
✓ Successfully reconstructed private share!
  Saved to: ./output/bafybmi.../pvtShare.png

============================================
Summary:
  Total DIDs: 3
  Successful: 3
  Failed: 0

IMPORTANT: Keep your private shares secure and never share them!
```

---

### 2. transfer

Transfer RBT tokens to another DID. Supports two modes: **Standard Mode** (uses environment variables) and **File Mode** (reads sender info from accounts.json).

#### Flags

**Common Flags (both modes):**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--receiver` | string | ✓ | Receiver DID |
| `--amount` | float64 | ✓ | Amount to transfer (must be > 0) |
| `--comment` | string | | Transfer comment/memo (optional) |

**File Mode Flags:**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--from-file` | string | ✓* | Path to accounts JSON file (exported via `export-dids`) |
| `--sender-index` | int | ✓* | Index of sender account in file (0-based) |

**Standard Mode Flags:**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--rubix-node` | string | | Rubix node URL (default: env `RUBIX_NODE_URL` or `localhost:20006`) |
| `--sender-did` | string | ✓* | Sender DID (default: env `SENDER_DID`) |
| `--preset` | string | | Preset folder path (default: env `PRESET_FOLDER` or `./preset`) |

*\* Required for standard mode (unless using file mode with `--from-file`)*

#### Examples

**Quick Transfer (Simplest - Assumes .env is configured):**

```bash
# Prerequisites: .env file with SENDER_DID, RUBIX_NODE_URL

# Step 1: Reconstruct private share for your sender DID
./break-nlss break-nlss --did bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy

# Step 2: Transfer tokens
./break-nlss transfer \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.5 \
  --comment "Payment for services"
```

**Direct Transfer with All Flags (No .env needed):**

```bash
# Step 1: Reconstruct private share
./break-nlss break-nlss --did bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy

# Step 2: Transfer with explicit flags
./break-nlss transfer \
  --rubix-node localhost:20006 \
  --sender-did bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.5 \
  --comment "Payment"
```

**File Mode (Recommended for Multiple Senders):**

```bash
# Step 1: Export DIDs with balances
./break-nlss export-dids --output accounts.json

# Step 2: Reconstruct private shares for all DIDs
./break-nlss break-nlss --did dids.txt

# Step 3: Transfer using account from file
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 5.0 \
  --comment "Payment from file"
```

#### Mode Comparison

| Aspect | File Mode | Standard Mode |
|--------|-----------|---------------|
| **Sender Config** | Read from JSON file | Environment vars or flags |
| **Balance Check** | ✅ Automatic validation | ❌ Manual |
| **Multiple Senders** | ✅ Change `--sender-index` | ❌ Change env vars each time |
| **Node Config** | ✅ Auto from file | Set via env/flag |
| **Use Case** | Multiple senders, batch operations | Single sender, frequent transfers |

#### Output

```
Transfer Configuration:
=======================
  Rubix Node: localhost:20006
  Sender DID: bafybmiguvjk...
  Receiver: bafybmiee3dmi...
  Amount: 5.00 RBT
  Sender Balance: 67.00 RBT
  Comment: Payment from file

Phase 1: Initiating transfer...
✓ Transaction initiated
  Request ID: txn_20251125_abc123
  Hash (Base64): SGVsbG8gV29ybGQh...

Phase 2: Generating signatures...
✓ Decoded hash: Hello World!
  Generating image signature from: ./output/bafybmiguvjk.../pvtShare.png
✓ Image signature generated (32 bytes)

Phase 3: Submitting signatures...

✓ Transaction completed successfully!
  Message: Transaction completed
```

---

### 3. balance

Query the RBT balance for a DID.

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--did` | string | | DID to query (default: env `SENDER_DID`) |
| `--rubix-node` | string | | Rubix node URL (default: env `RUBIX_NODE_URL` or `localhost:20006`) |

#### Examples

**Check your own balance (uses SENDER_DID from env):**
```bash
export SENDER_DID="bafybmi..."
./break-nlss balance
```

**Check another DID's balance:**
```bash
./break-nlss balance --did bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
```

**With specific Rubix node:**
```bash
./break-nlss balance \
  --rubix-node localhost:20006 \
  --did bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
```

#### Output

```
Querying balance for DID: bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
Rubix Node: localhost:20006

Balance: 67.00 RBT
```

---

### 4. list-dids

List all DIDs registered on the Rubix node with their balances.

#### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--rubix-node` | string | | Rubix node URL (default: env `RUBIX_NODE_URL` or `localhost:20006`) |

#### Examples

```bash
# Use default node from env
./break-nlss list-dids

# With specific node
./break-nlss list-dids --rubix-node localhost:20006
```

#### Output

```
Fetching all DIDs from: localhost:20006

Status: true
Message: Got all DID
Total DIDs: 5

Account Information:
====================

[1] DID: bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy
    Type: 4
    RBT Amount: 67.00
    Pledged: 0.00 | Locked: 0.90 | Pinned: 0.00

[2] DID: bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
    Type: 4
    RBT Amount: 45.50
    Pledged: 0.00 | Locked: 0.00 | Pinned: 0.00
...
```

---

### 5. export-dids

Export DIDs with balance > 0 (or custom threshold) to a JSON file for file-based transfers.

#### Flags

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--output` | string | | `accounts.json` | Output file path |
| `--rubix-node` | string | | env or `localhost:20006` | Rubix node URL |
| `--min-balance` | float64 | | `0.0` | Minimum balance to include (0 = only non-zero balances) |

#### Examples

**Export all DIDs with balance > 0:**
```bash
./break-nlss export-dids --output accounts.json
```

**Export DIDs with balance > 10 RBT:**
```bash
./break-nlss export-dids --output accounts.json --min-balance 10.0
```

**Export from specific node:**
```bash
./break-nlss export-dids \
  --output my-accounts.json \
  --rubix-node localhost:8080 \
  --min-balance 5.0
```

#### Output File Format

**accounts.json:**
```json
{
  "version": "1.0",
  "rubix_node_url": "localhost:20006",
  "exported_at": "2025-11-25T18:53:39.111233+05:30",
  "total_dids": 3,
  "accounts": [
    {
      "did": "bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy",
      "balance": 67.0,
      "did_type": 4,
      "pledged_rbt": 0,
      "locked_rbt": 0.9,
      "pinned_rbt": 0,
      "updated_at": "2025-11-25T18:53:39.111228+05:30"
    }
  ]
}
```

#### Console Output

```
Fetching DIDs from: localhost:20006
Minimum balance filter: 0.00 RBT

Total DIDs on node: 10
DIDs with balance > 0.00: 3

✓ Successfully exported 3 DIDs to: accounts.json

Exported Accounts:
==================
[0] DID: bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy
    Balance: 67.00 RBT
    Pledged: 0.00 | Locked: 0.90 | Pinned: 0.00

[1] DID: bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u
    Balance: 45.50 RBT
    Pledged: 0.00 | Locked: 0.00 | Pinned: 0.00

Usage:
  ./break-nlss transfer --from-file accounts.json --sender-index 0 --receiver <DID> --amount <AMOUNT>
```

---

### 6. generate-key

Generate a new EC key pair (P-256 curve) for ECDSA operations. **Note:** Currently not required for token transfers.

#### Flags

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--output` | string | | `./preset` | Output directory for key files |

#### Examples

```bash
# Generate keys in default location
./break-nlss generate-key

# Generate keys in custom directory
./break-nlss generate-key --output ./keys
```

#### Output

Creates two files:
- `{output}/privatekey.pem` - Private key (KEEP SECURE!)
- `{output}/publickey.pem` - Public key

```
Generating new EC key pair (P-256)...
Output directory: ./preset

✓ Key pair generated successfully!
  Private key: ./preset/privatekey.pem
  Public key: ./preset/publickey.pem

IMPORTANT: Keep your private key secure and never share it!
```

---

### 7. help

Display help information about available commands.

#### Examples

```bash
./break-nlss help
./break-nlss --help
./break-nlss
```

#### Output

```
Break-NLSS - Rubix Blockchain Token Transfer Tool
Version: 1.0.0

Usage:
  break-nlss <command> [options]

Commands:
  transfer     - Transfer tokens to another DID
  balance      - Get account balance for a DID
  list-dids    - List all DIDs from the node
  export-dids  - Export DIDs with balance > 0 to a file
  generate-key - Generate a new EC key pair
  break-nlss   - Reconstruct private share from DID and public share
  help         - Show this help message

Environment Variables:
  RUBIX_NODE_URL  - Rubix node URL (default: localhost:20006)
  SENDER_PEER_ID  - Sender peer ID
  SENDER_DID      - Sender DID
  PRESET_FOLDER   - Path to preset folder (default: ./preset)

Examples:
  # Export DIDs with balance > 0 to file
  break-nlss export-dids --output accounts.json

  # Transfer tokens from file
  break-nlss transfer --from-file accounts.json --sender-index 0 --receiver bafybmi... --amount 10.5

  # Get balance
  break-nlss balance --did bafybmi...
```

---

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Rubix Node Configuration
RUBIX_NODE_URL=localhost:20006
SENDER_DID=bafybmi...

# NLSS Configuration
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1
NLSS_OUTPUT_DIR=./output

# Optional
PRESET_FOLDER=./preset
```

### Configuration Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RUBIX_NODE_URL` | Rubix blockchain node URL | `localhost:20006` |
| `SENDER_DID` | Your DID | (required for transfers) |
| `NLSS_BASE_PATH` | Base path for Rubix data directory | (required for break-nlss) |
| `NLSS_NODE_NAME` | Rubix node name | (required for break-nlss) |
| `NLSS_OUTPUT_DIR` | Output directory for pvtShare.png | `./output` |
| `PRESET_FOLDER` | Path to preset folder | `./preset` |

### .env.example

```bash
# Rubix Node Configuration
RUBIX_NODE_URL=localhost:20006
SENDER_DID=your_did_here

# NLSS Configuration (for break-nlss command)
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1
NLSS_DID_IMAGE_NAME=DIDImg.png
NLSS_PUB_SHARE_NAME=pubShare.png
NLSS_OUTPUT_DIR=./output

# Optional
PRESET_FOLDER=./preset
```

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Access to a Rubix blockchain node
- Rubix DID with associated DID image and public share

### Build from Source

```bash
# Clone or navigate to the repository
cd /Users/allen/Professional/break-nlss

# Install dependencies
go mod download

# Build the application
go build -o break-nlss

# Verify build
./break-nlss help
```

---

## Complete Workflow Examples

### Workflow 1: Simple Single Sender Transfer

The simplest workflow for a single sender doing regular transfers.

```bash
# Step 1: Configure environment
cat > .env <<EOF
RUBIX_NODE_URL=localhost:20006
SENDER_DID=bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1
NLSS_OUTPUT_DIR=./output
EOF

# Step 2: Check your balance
./break-nlss balance

# Step 3: Reconstruct private share (one-time setup)
./break-nlss break-nlss --did bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy

# Step 4: Transfer tokens (repeat as needed)
./break-nlss transfer \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.0 \
  --comment "Payment for services"

# Step 5: Verify balance after transfer
./break-nlss balance
```

### Workflow 2: File-Based Transfer (Multiple Senders)

Recommended for managing multiple senders and batch operations.

```bash
# Step 1: Configure environment
cat > .env <<EOF
RUBIX_NODE_URL=localhost:20006
NLSS_BASE_PATH=/Users/allen/Professional/sky
NLSS_NODE_NAME=node1
NLSS_OUTPUT_DIR=./output
EOF

# Step 2: Export all DIDs with balances
./break-nlss export-dids --output accounts.json

# Step 3: Create DIDs list file
cat accounts.json | jq -r '.accounts[].did' > dids.txt

# Step 4: Reconstruct private shares for all DIDs
./break-nlss break-nlss --did dids.txt

# Step 5: Perform transfer
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u \
  --amount 10.0 \
  --comment "Test transfer"
```

### Workflow 3: Batch Processing Multiple DIDs

Process and prepare multiple DIDs at once.

```bash
# Step 1: Get all DIDs from node
./break-nlss list-dids > all_dids.txt

# Step 2: Extract DIDs to file
grep "DID:" all_dids.txt | awk '{print $3}' > dids.txt

# Step 3: Batch reconstruct private shares
./break-nlss break-nlss --did dids.txt

# Step 4: Export to accounts file
./break-nlss export-dids --output accounts.json

# Step 5: Now you can use any account for transfers
./break-nlss transfer \
  --from-file accounts.json \
  --sender-index 0 \
  --receiver <RECEIVER_DID> \
  --amount 5.0
```

---

## Architecture

### Project Structure

```
break-nlss/
├── main.go                 # CLI entry point and command handlers
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── .env                    # Environment configuration
├── .gitignore              # Git ignore rules
│
├── pkg/                    # Public packages
│   ├── config/             # Configuration management
│   │   └── config.go       # Config loading, validation, path construction
│   │
│   ├── crypto/             # Cryptographic operations
│   │   ├── hash.go         # SHA3-256 hashing (currently unused)
│   │   ├── ecdsa.go        # ECDSA key operations (currently unused)
│   │   └── image.go        # Image-based signature generation
│   │
│   ├── nlss/               # NLSS algorithm implementation
│   │   └── nlss.go         # Break-NLSS reconstruction, verification, signing
│   │
│   ├── rubix/              # Rubix blockchain client
│   │   ├── client.go       # HTTP client wrapper
│   │   ├── transaction.go  # Token transfer operations
│   │   └── models.go       # Request/Response structs
│   │
│   └── storage/            # File-based storage
│       └── accounts.go     # DID account persistence (JSON)
│
├── internal/               # Private packages
│   └── files/              # File loading utilities
│       └── loader.go       # Generic file loader
│
├── output/                 # Generated private shares (gitignored)
│   └── {DID}/
│       └── pvtShare.png
│
├── preset/                 # User-provided files (gitignored)
│   ├── privatekey.pem      # EC private key (currently unused)
│   └── (other user files)
│
└── accounts.json           # Exported DID accounts (gitignored)
```

### Package Descriptions

#### pkg/config
- Loads configuration from environment variables and .env file
- Constructs dynamic paths for NLSS operations
- Validates required configuration
- Provides helpers: `GetNLSSImagePaths()`, `GetNLSSOutputPath()`

#### pkg/crypto
- **image.go**: Image-based signature generation from PNG files
  - `GetPNGImagePixels()`: Extract pixel data as bytes
  - `RandomPositions()`: Generate deterministic bit positions from hash
  - `Sign()`: Create 32-byte signature from private share image
- **hash.go**: SHA3-256 hashing (currently not used in transfer flow)
- **ecdsa.go**: ECDSA key operations (currently not used in transfer flow)

#### pkg/nlss
- **Break-NLSS Algorithm**: Reconstructs private share from DID + public share
- **Key functions:**
  - `BreakNLSS()`: Core algorithm (XOR-based reconstruction)
  - `BreakNLSSFromFiles()`: File-based wrapper
  - `VerifyPVT()`: Cryptographic verification of reconstructed share
  - `Sign()`: Generate signature from private share (wrapper)
  - `RandomPositions()`: Deterministic position generation

#### pkg/rubix
- **client.go**: HTTP client for Rubix blockchain REST APIs
- **transaction.go**: Two-phase token transfer implementation
  - Phase 1: Initiate transfer (get transaction ID + hash)
  - Phase 2: Generate image signature and submit
- **models.go**: Request/response structs for all API calls

#### pkg/storage
- **accounts.go**: File-based account management
- Exports DIDs with balances to JSON
- Loads accounts for file-based transfers
- Supports account lookup by index

---

## How It Works

### Token Transfer Flow

The application performs a **two-phase token transfer** with **image-based signatures only** (ECDSA signing removed):

#### Phase 1: Initiate Transfer
1. Send `POST /api/initiate-rbt-transfer` with:
   - Sender DID (just the DID, no peer ID prefix needed)
   - Receiver DID
   - Amount
   - Comment
   - Type: 2 (RBT transfer)

2. Receive response containing:
   - Transaction ID (`id`)
   - Base64-encoded hash to sign

#### Phase 2: Generate Signature and Submit
1. **Decode hash**: Convert Base64 to string
2. **Load private share**: Read `pvtShare.png` from `output/{sender_did}/pvtShare.png`
3. **Generate image signature**:
   - Extract pixel data from PNG (RGB values → binary string)
   - Use hash to generate 256 deterministic bit positions
   - Extract bits at those positions
   - Pack 256 bits into 32 bytes
4. **Submit signatures** via `POST /api/signature-response`:
   - **Signature**: Empty byte array `[]` (ECDSA not used)
   - **Pixels**: 32-byte image signature

### Break-NLSS Algorithm

The Break-NLSS algorithm reconstructs the private share from DID and public share images:

#### Algorithm Steps

1. **Load Images**
   - DID Image: `{base_path}/{node_name}/Rubix/{DID}/DIDImg.png`
   - Public Share: `{base_path}/{node_name}/Rubix/{DID}/pubShare.png`

2. **Convert to Bytes**
   - Extract RGB pixel values
   - Convert to byte arrays

3. **Reconstruct Private Share**
   - XOR operation: `pvtShare = DID ⊕ pubShare`
   - Results in private share bytes

4. **Verify Reconstruction**
   - Cryptographic verification to ensure correctness
   - `VerifyPVT(did, pub, pvt)` returns true/false

5. **Save to PNG**
   - Create PNG image from private share bytes
   - Save to: `{output_dir}/{DID}/pvtShare.png`

### Image-Based Signature Generation

This algorithm must match the Dart reference implementation exactly:

#### Signature Algorithm

```
1. Convert PNG to Binary String
   - For each pixel: RGB → binary string
   - Concatenate all pixel binaries

2. Generate 256 Deterministic Bit Positions
   - For each character in hash (32 chars):
     for k in range(8):
       position = (((2402 + hashChar) * 2709) + ((k + 2709) + hashChar)) % 2048

3. Extract Signature Bits
   - Get bit at each position from image binary
   - Collect 256 bits

4. Pack to Bytes
   - Convert 256 bits → 32 bytes
   - Return byte array
```

**Note:** The magic numbers (2402, 2709, 2048) are from the original NLSS specification.

---

## API Reference

### Rubix Blockchain REST APIs

#### 1. Initiate RBT Transfer

**Endpoint:** `POST /api/initiate-rbt-transfer`

**Request:**
```json
{
  "receiver": "bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u",
  "sender": "bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy",
  "tokenCount": 10.5,
  "comment": "Payment for services",
  "type": 2
}
```

**Note:** The sender field now only requires the DID (no peer ID prefix needed).

---

## Architecture

### Project Structure

```
break-nlss/
├── main.go                 # CLI entry point and command handlers
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── .env                    # Environment configuration
├── .gitignore              # Git ignore rules
│
├── pkg/                    # Public packages
│   ├── config/             # Configuration management
│   │   └── config.go       # Config loading, validation, path construction
│   │
│   ├── crypto/             # Cryptographic operations
│   │   ├── hash.go         # SHA3-256 hashing (currently unused)
│   │   ├── ecdsa.go        # ECDSA key operations (currently unused)
│   │   └── image.go        # Image-based signature generation
│   │
│   ├── nlss/               # NLSS algorithm implementation
│   │   └── nlss.go         # Break-NLSS reconstruction, verification, signing
│   │
│   ├── rubix/              # Rubix blockchain client
│   │   ├── client.go       # HTTP client wrapper
│   │   ├── transaction.go  # Token transfer operations
│   │   └── models.go       # Request/Response structs
│   │
│   └── storage/            # File-based storage
│       └── accounts.go     # DID account persistence (JSON)
│
├── internal/               # Private packages
│   └── files/              # File loading utilities
│       └── loader.go       # Generic file loader
│
├── output/                 # Generated private shares (gitignored)
│   └── {DID}/
│       └── pvtShare.png
│
├── preset/                 # User-provided files (gitignored)
│   ├── privatekey.pem      # EC private key (currently unused)
│   └── (other user files)
│
└── accounts.json           # Exported DID accounts (gitignored)
```

### Package Descriptions

#### pkg/config
- Loads configuration from environment variables and .env file
- Constructs dynamic paths for NLSS operations
- Validates required configuration
- Provides helpers: `GetNLSSImagePaths()`, `GetNLSSOutputPath()`

#### pkg/crypto
- **image.go**: Image-based signature generation from PNG files
  - `GetPNGImagePixels()`: Extract pixel data as bytes
  - `RandomPositions()`: Generate deterministic bit positions from hash
  - `Sign()`: Create 32-byte signature from private share image
- **hash.go**: SHA3-256 hashing (currently not used in transfer flow)
- **ecdsa.go**: ECDSA key operations (currently not used in transfer flow)

#### pkg/nlss
- **Break-NLSS Algorithm**: Reconstructs private share from DID + public share
- **Key functions:**
  - `BreakNLSS()`: Core algorithm (XOR-based reconstruction)
  - `BreakNLSSFromFiles()`: File-based wrapper
  - `VerifyPVT()`: Cryptographic verification of reconstructed share
  - `Sign()`: Generate signature from private share (wrapper)
  - `RandomPositions()`: Deterministic position generation

#### pkg/rubix
- **client.go**: HTTP client for Rubix blockchain REST APIs
- **transaction.go**: Two-phase token transfer implementation
  - Phase 1: Initiate transfer (get transaction ID + hash)
  - Phase 2: Generate image signature and submit
- **models.go**: Request/response structs for all API calls

#### pkg/storage
- **accounts.go**: File-based account management
- Exports DIDs with balances to JSON
- Loads accounts for file-based transfers
- Supports account lookup by index

---

## How It Works

### Token Transfer Flow

The application performs a **two-phase token transfer** with **image-based signatures only** (ECDSA signing removed):

#### Phase 1: Initiate Transfer
1. Send `POST /api/initiate-rbt-transfer` with:
   - Sender DID (with peer ID prefix: `{peer_id}.{did}`)
   - Receiver DID
   - Amount
   - Comment
   - Type: 2 (RBT transfer)

2. Receive response containing:
   - Transaction ID (`id`)
   - Base64-encoded hash to sign

#### Phase 2: Generate Signature and Submit
1. **Decode hash**: Convert Base64 to string
2. **Load private share**: Read `pvtShare.png` from `output/{sender_did}/pvtShare.png`
3. **Generate image signature**:
   - Extract pixel data from PNG (RGB values → binary string)
   - Use hash to generate 256 deterministic bit positions
   - Extract bits at those positions
   - Pack 256 bits into 32 bytes
4. **Submit signatures** via `POST /api/signature-response`:
   - **Signature**: Empty byte array `[]` (ECDSA not used)
   - **Pixels**: 32-byte image signature

### Break-NLSS Algorithm

The Break-NLSS algorithm reconstructs the private share from DID and public share images:

#### Algorithm Steps

1. **Load Images**
   - DID Image: `{base_path}/{node_name}/Rubix/{DID}/DIDImg.png`
   - Public Share: `{base_path}/{node_name}/Rubix/{DID}/pubShare.png`

2. **Convert to Bytes**
   - Extract RGB pixel values
   - Convert to byte arrays

3. **Reconstruct Private Share**
   - XOR operation: `pvtShare = DID ⊕ pubShare`
   - Results in private share bytes

4. **Verify Reconstruction**
   - Cryptographic verification to ensure correctness
   - `VerifyPVT(did, pub, pvt)` returns true/false

5. **Save to PNG**
   - Create PNG image from private share bytes
   - Save to: `{output_dir}/{DID}/pvtShare.png`

### Image-Based Signature Generation

This algorithm must match the Dart reference implementation exactly:

#### Signature Algorithm

```
1. Convert PNG to Binary String
   - For each pixel: RGB → binary string
   - Concatenate all pixel binaries

2. Generate 256 Deterministic Bit Positions
   - For each character in hash (32 chars):
     for k in range(8):
       position = (((2402 + hashChar) * 2709) + ((k + 2709) + hashChar)) % 2048

3. Extract Signature Bits
   - Get bit at each position from image binary
   - Collect 256 bits

4. Pack to Bytes
   - Convert 256 bits → 32 bytes
   - Return byte array
```

**Note:** The magic numbers (2402, 2709, 2048) are from the original NLSS specification.

---

## API Reference

### Rubix Blockchain REST APIs

#### 1. Initiate RBT Transfer

**Endpoint:** `POST /api/initiate-rbt-transfer`

**Request:**
```json
{
  "receiver": "bafybmiee3dmi25jxpev4rwjli23yxndihsqcayvtxwy2pa6vz4qs2no64u",
  "sender": "12D3Koo....bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy",
  "tokenCount": 10.5,
  "comment": "Payment for services",
  "type": 2
}
```

⚠️ **Note:** Field name is `tokenCount` (camelCase with capital 'C')

**Response:**
```json
{
  "status": true,
  "result": {
    "id": "txn_20251125_abc123",
    "hash": "SGVsbG8gV29ybGQh..."
  },
  "message": "Transaction initiated successfully"
}
```

#### 2. Submit Signature

**Endpoint:** `POST /api/signature-response`

**Request:**
```json
{
  "id": "txn_20251125_abc123",
  "signature": {
    "Signature": [],
    "Pixels": [1, 0, 1, 0, 1, 1, 0, 1, ...]
  }
}
```

**Response:**
```json
{
  "status": true,
  "message": "Transaction completed successfully"
}
```

#### 3. Get Account Balance

**Endpoint:** `GET /api/get-account-info?did={did}`

**Response:**
```json
{
  "status": true,
  "account_info": [{
    "did": "bafybmiguvjk5nqxjmrdhfna42dzpgloy47d7r3vncsax6nxe3irir4vkdy",
    "did_type": 4,
    "rbt_amount": 67.0,
    "pledged_rbt": 0,
    "locked_rbt": 0.9,
    "pinned_rbt": 0
  }]
}
```

#### 4. Get All DIDs

**Endpoint:** `GET /api/get-all-did`

**Response:**
```json
{
  "status": true,
  "message": "Got all DID",
  "account_info": [
    {
      "did": "bafybmi...",
      "did_type": 4,
      "rbt_amount": 67.0,
      "pledged_rbt": 0,
      "locked_rbt": 0.9,
      "pinned_rbt": 0
    }
  ]
}
```

---

## Troubleshooting

### Common Issues

#### 1. break-nlss Command Errors

**Error:** "Error: --did is required"
- **Solution:** Provide `--did` flag with either a DID string or path to file containing DIDs

**Error:** "DID image file not found"
- **Solution:** Ensure the DID image exists at: `{NLSS_BASE_PATH}/{NLSS_NODE_NAME}/Rubix/{DID}/DIDImg.png`
- Check that `NLSS_BASE_PATH` and `NLSS_NODE_NAME` are correctly set in `.env`

**Error:** "Public share file not found"
- **Solution:** Ensure public share exists at: `{NLSS_BASE_PATH}/{NLSS_NODE_NAME}/Rubix/{DID}/pubShare.png`

**Error:** "Verification: false"
- **Solution:** The reconstructed private share failed verification. Ensure:
  - DID image is correct
  - Public share is correct
  - Files are not corrupted

#### 2. Transfer Command Errors

**Error:** "Configuration error: SENDER_PEER_ID is required"
- **Solution Option 1:** Use file mode: `--from-file accounts.json --sender-index 0`
- **Solution Option 2:** Set `SENDER_PEER_ID` environment variable
- **Solution Option 3:** Use `--sender-peer` flag

**Error:** "failed to generate image signature"
- **Solution:** Run `break-nlss` command first to generate private share:
  ```bash
  ./break-nlss break-nlss --did {SENDER_DID}
  ```

**Error:** "Insufficient balance"
- **Solution:** Check balance and ensure sender has enough RBT:
  ```bash
  ./break-nlss balance --did {SENDER_DID}
  ```

**Error:** "Invalid sender index"
- **Solution:** Check accounts file to see available accounts:
  ```bash
  cat accounts.json | jq '.accounts | length'
  ```
  Remember: indexes are 0-based (first account = index 0)

#### 3. Configuration Errors

**Error:** "Error loading config: failed to load .env"
- **Solution:** Create `.env` file in project root or set environment variables manually

**Error:** "NLSS_BASE_PATH is required"
- **Solution:** Set `NLSS_BASE_PATH` in `.env`:
  ```bash
  NLSS_BASE_PATH=/Users/allen/Professional/sky
  ```

#### 4. File Format Errors

**Error:** "Error loading accounts file: invalid JSON"
- **Solution:** Validate JSON format:
  ```bash
  cat accounts.json | jq .
  ```
- Re-export if corrupted:
  ```bash
  ./break-nlss export-dids --output accounts.json
  ```

**Error:** "Error reading DIDs from file"
- **Solution:** Ensure dids.txt has one DID per line:
  ```
  bafybmi...
  bafybmi...
  ```

#### 5. API Errors

**Error:** "failed to initiate transfer: connection refused"
- **Solution:**
  - Ensure Rubix node is running
  - Check `RUBIX_NODE_URL` is correct
  - Try: `curl http://{RUBIX_NODE_URL}/api/get-all-did`

**Error:** "signature submission failed: unauthorized"
- **Solution:**
  - Ensure private share is correct for the sender DID
  - Re-run `break-nlss` command to regenerate private share

### Debug Mode

For detailed debugging information:

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o break-nlss-debug

# Run with verbose output (if implemented)
./break-nlss-debug transfer --receiver ... --amount ... -v
```

---

## Security Considerations

### File Security

1. **Private Shares (pvtShare.png)**
   - Generated in `output/{DID}/` directory
   - **CRITICAL:** Never share or commit these files
   - Add `output/` to `.gitignore`
   - Protect with filesystem permissions: `chmod 700 output/`

2. **Private Keys (privatekey.pem)**
   - Currently not used for transfers but may be required for other operations
   - Store with restricted permissions: `chmod 600 preset/privatekey.pem`
   - Never commit to version control
   - Add `preset/` to `.gitignore`

3. **Accounts File (accounts.json)**
   - Contains DID information and balances
   - May contain peer IDs (not highly sensitive but should be protected)
   - Add to `.gitignore`

### Network Security

1. **HTTPS Usage**
   - Use HTTPS when connecting to remote Rubix nodes
   - Validate SSL certificates in production

2. **Environment Variables**
   - Never commit `.env` file to version control
   - Use `.env.example` as template
   - Restrict `.env` permissions: `chmod 600 .env`

### Recommended .gitignore

```gitignore
# Environment
.env

# Build outputs
break-nlss
*.exe

# Sensitive files
preset/
output/
accounts.json
*.pem
*.png

# Go
*.log
*.test
*.out
vendor/
```

---

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./pkg/nlss/...

# Run with coverage
go test -cover ./pkg/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Development

### Adding New Features

1. Identify the appropriate package (config, crypto, nlss, rubix, storage)
2. Implement functionality with proper error handling
3. Write unit tests
4. Update documentation
5. Test with real Rubix node

### Code Style

- Follow Go conventions and idioms
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly (no panic in library code)
- Use structured logging where appropriate

---

## References

### Related Projects
- **Rubix Blockchain**: https://rubix.net
- **Sky Server**: `/Users/allen/Professional/sky`
- **fexr-flutter SDK**: `/Users/allen/Professional/sky/fexr-flutter`

### Dart Reference Implementations
- Image processing: `lib/signature/dependencies.dart`
- Signature generation: `lib/signature/gen_sign.dart`
- ECDSA operations: `lib/signature/key_gen.dart`
- Rubix APIs: `lib/native_interaction/rubix/rubix_platform_calls.dart`

---

## License

[Your License Here]

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

---

## Support

For issues and questions:
- Check the [Troubleshooting](#troubleshooting) section
- Review the [API Reference](#api-reference)
- Check Rubix blockchain documentation

---

**Version:** 1.0.0
**Last Updated:** 2025-11-25
