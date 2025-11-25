package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"break-nlss/pkg/config"
	"break-nlss/pkg/nlss"
	"break-nlss/pkg/rubix"
	"break-nlss/pkg/storage"
)

const (
	version = "1.0.0"
)

func printUsage() {
	fmt.Println("Break-NLSS - Rubix Blockchain Token Transfer Tool")
	fmt.Printf("Version: %s\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  break-nlss <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  transfer     - Transfer tokens to another DID")
	fmt.Println("  balance      - Get account balance for a DID")
	fmt.Println("  list-dids    - List all DIDs from the node")
	fmt.Println("  export-dids  - Export DIDs with balance > 0 to a file")
	fmt.Println("  generate-key - Generate a new EC key pair")
	fmt.Println("  break-nlss   - Reconstruct private share from DID and public share")
	fmt.Println("  help         - Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  RUBIX_NODE_URL  - Rubix node URL (default: localhost:20006)")
	fmt.Println("  SENDER_PEER_ID  - Sender peer ID")
	fmt.Println("  SENDER_DID      - Sender DID")
	fmt.Println("  PRESET_FOLDER   - Path to preset folder (default: ./preset)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Export DIDs with balance > 0 to file")
	fmt.Println("  break-nlss export-dids --output accounts.json")
	fmt.Println()
	fmt.Println("  # Transfer tokens from file")
	fmt.Println("  break-nlss transfer --from-file accounts.json --sender-index 0 --receiver bafybmi... --amount 10.5")
	fmt.Println()
	fmt.Println("  # Get balance")
	fmt.Println("  break-nlss balance --did bafybmi...")
	fmt.Println()
	fmt.Println("  # Generate new keys")
	fmt.Println("  break-nlss generate-key --output ./preset")
	fmt.Println()
	fmt.Println("  # Reconstruct private share from single DID")
	fmt.Println("  break-nlss break-nlss --did bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse")
	fmt.Println()
	fmt.Println("  # Reconstruct private shares from multiple DIDs in file")
	fmt.Println("  break-nlss break-nlss --did dids.txt")
	fmt.Println()
}

func main() {
	// Load .env file if it exists (ignore error if file doesn't exist)
	godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "transfer":
		runTransfer()
	case "balance":
		runBalance()
	case "list-dids":
		runListDIDs()
	case "export-dids":
		runExportDIDs()
	case "generate-key":
		runGenerateKey()
	case "break-nlss":
		runBreakNLSS()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runTransfer() {
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)

	// Standard mode flags
	receiver := transferCmd.String("receiver", "", "Receiver DID (required)")
	amount := transferCmd.Float64("amount", 0, "Amount to transfer (required)")
	comment := transferCmd.String("comment", "", "Transfer comment (optional)")
	rubixNode := transferCmd.String("rubix-node", "", "Rubix node URL (default: from env or localhost:20006)")
	presetFolder := transferCmd.String("preset", "", "Preset folder path (default: from env or ./preset)")
	senderPeerID := transferCmd.String("sender-peer", "", "Sender peer ID (default: from env)")
	senderDID := transferCmd.String("sender-did", "", "Sender DID (default: from env)")

	// File mode flags
	fromFile := transferCmd.String("from-file", "", "Read sender info from accounts file")
	senderIndex := transferCmd.Int("sender-index", -1, "Index of sender in accounts file (0-based)")

	transferCmd.Parse(os.Args[2:])

	// Validate required flags
	if *receiver == "" {
		fmt.Println("Error: --receiver is required")
		transferCmd.Usage()
		os.Exit(1)
	}

	if *amount <= 0 {
		fmt.Println("Error: --amount must be greater than 0")
		transferCmd.Usage()
		os.Exit(1)
	}

	var finalSenderPeerID, finalSenderDID string
	var senderBalance float64

	// Check if using file mode
	if *fromFile != "" {
		if *senderIndex < 0 {
			fmt.Println("Error: --sender-index is required when using --from-file")
			transferCmd.Usage()
			os.Exit(1)
		}

		// Load accounts from file
		accountsFile, err := storage.LoadAccountsFromFile(*fromFile)
		if err != nil {
			fmt.Printf("Error loading accounts file: %v\n", err)
			os.Exit(1)
		}

		// Get sender account by index
		sender := accountsFile.GetAccountByIndex(*senderIndex)
		if sender == nil {
			fmt.Printf("Error: Invalid sender index %d. File has %d accounts.\n", *senderIndex, len(accountsFile.Accounts))
			os.Exit(1)
		}

		finalSenderDID = sender.DID
		finalSenderPeerID = sender.PeerID
		senderBalance = sender.Balance

		fmt.Printf("Using sender from file: %s (Balance: %.2f RBT)\n", finalSenderDID, senderBalance)

		// Check if sender has enough balance
		if senderBalance < *amount {
			fmt.Printf("Error: Insufficient balance. Sender has %.2f RBT, trying to send %.2f RBT\n", senderBalance, *amount)
			os.Exit(1)
		}

		// Use Rubix node URL from file if not overridden
		if *rubixNode == "" {
			*rubixNode = accountsFile.RubixNodeURL
		}
	} else {
		// Standard mode - use flags or env
		finalSenderPeerID = *senderPeerID
		finalSenderDID = *senderDID
	}

	// Load configuration
	cfg, err := config.LoadConfigWithOverrides(*rubixNode, *presetFolder, finalSenderPeerID, finalSenderDID)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		fmt.Println("\nPlease set the following environment variables:")
		fmt.Println("  RUBIX_NODE_URL  - Rubix node URL")
		fmt.Println("  SENDER_PEER_ID  - Your peer ID")
		fmt.Println("  SENDER_DID      - Your DID")
		fmt.Println("\nOr use command-line flags:")
		fmt.Println("  --from-file accounts.json --sender-index 0")
		transferCmd.Usage()
		os.Exit(1)
	}

	// Print configuration
	fmt.Println("\nTransfer Configuration:")
	fmt.Println("=======================")
	cfg.PrintConfig()
	fmt.Printf("  Receiver: %s\n", *receiver)
	fmt.Printf("  Amount: %.2f RBT\n", *amount)
	if *fromFile != "" {
		fmt.Printf("  Sender Balance: %.2f RBT\n", senderBalance)
	}
	fmt.Printf("  Comment: %s\n", *comment)
	fmt.Println()

	// Perform transfer
	params := rubix.TransferParams{
		RubixNodeURL:     cfg.RubixNodeURL,
		SenderPeerID:     cfg.SenderPeerID,
		SenderDID:        cfg.SenderDID,
		ReceiverDID:      *receiver,
		Amount:           *amount,
		Comment:          *comment,
		PrivateKeyPath:   cfg.PrivateKeyPath,
		PrivateSharePath: cfg.PrivateSharePath,
		NLSSOutputDir:    cfg.NLSSOutputDir,
	}

	if err := rubix.TransferTokens(params); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
}

func runBalance() {
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)

	did := balanceCmd.String("did", "", "DID to query (default: from env SENDER_DID)")
	rubixNode := balanceCmd.String("rubix-node", "", "Rubix node URL (default: from env or localhost:20006)")

	balanceCmd.Parse(os.Args[2:])

	// Load configuration
	cfg, err := config.LoadConfigWithOverrides(*rubixNode, "", "", *did)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Use SENDER_DID if --did not provided
	queryDID := *did
	if queryDID == "" {
		queryDID = cfg.SenderDID
	}

	if queryDID == "" {
		fmt.Println("Error: --did is required or set SENDER_DID environment variable")
		balanceCmd.Usage()
		os.Exit(1)
	}

	fmt.Printf("Querying balance for DID: %s\n", queryDID)
	fmt.Printf("Rubix Node: %s\n\n", cfg.RubixNodeURL)

	// Get balance
	balance, err := rubix.GetAccountBalance(cfg.RubixNodeURL, queryDID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Balance: %.2f RBT\n", balance)
}

func runListDIDs() {
	listCmd := flag.NewFlagSet("list-dids", flag.ExitOnError)

	rubixNode := listCmd.String("rubix-node", "", "Rubix node URL (default: from env or localhost:20006)")

	listCmd.Parse(os.Args[2:])

	// Load configuration
	cfg, err := config.LoadConfigWithOverrides(*rubixNode, "", "", "")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching all DIDs from: %s\n\n", cfg.RubixNodeURL)

	// Get all DIDs
	client := rubix.NewClient(cfg.RubixNodeURL)
	response, err := client.GetAllDID()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status: %v\n", response.Status)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Total DIDs: %d\n\n", len(response.AccountInfo))

	fmt.Println("Account Information:")
	fmt.Println("====================")
	for i, account := range response.AccountInfo {
		fmt.Printf("\n[%d] DID: %s\n", i+1, account.DID)
		fmt.Printf("    Type: %d\n", account.DIDType)
		fmt.Printf("    RBT Amount: %.2f\n", account.RBTAmount)
		fmt.Printf("    Pledged: %.2f | Locked: %.2f | Pinned: %.2f\n",
			account.PledgedRBT, account.LockedRBT, account.PinnedRBT)
	}
}

func runExportDIDs() {
	exportCmd := flag.NewFlagSet("export-dids", flag.ExitOnError)

	output := exportCmd.String("output", "accounts.json", "Output file path")
	rubixNode := exportCmd.String("rubix-node", "", "Rubix node URL (default: from env or localhost:20006)")
	minBalance := exportCmd.Float64("min-balance", 0.0, "Minimum balance to include (default: 0, only non-zero balances)")

	exportCmd.Parse(os.Args[2:])

	// Load configuration
	cfg, err := config.LoadConfigWithOverrides(*rubixNode, "", "", "")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching DIDs from: %s\n", cfg.RubixNodeURL)
	fmt.Printf("Minimum balance filter: %.2f RBT\n\n", *minBalance)

	// Get all DIDs
	client := rubix.NewClient(cfg.RubixNodeURL)
	response, err := client.GetAllDID()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total DIDs on node: %d\n", len(response.AccountInfo))

	// Filter DIDs with balance > minBalance
	var accounts []storage.DIDAccount
	for _, account := range response.AccountInfo {
		if account.RBTAmount > *minBalance {
			accounts = append(accounts, storage.DIDAccount{
				DID:        account.DID,
				PeerID:     "", // Not available from API, will need to be set manually if needed
				Balance:    account.RBTAmount,
				DIDType:    account.DIDType,
				PledgedRBT: account.PledgedRBT,
				LockedRBT:  account.LockedRBT,
				PinnedRBT:  account.PinnedRBT,
				UpdatedAt:  time.Now(),
			})
		}
	}

	fmt.Printf("DIDs with balance > %.2f: %d\n\n", *minBalance, len(accounts))

	if len(accounts) == 0 {
		fmt.Println("No DIDs found with the specified balance criteria.")
		os.Exit(0)
	}

	// Save to file
	err = storage.SaveAccountsToFile(*output, accounts, cfg.RubixNodeURL)
	if err != nil {
		fmt.Printf("Error saving to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully exported %d DIDs to: %s\n\n", len(accounts), *output)

	// Print summary
	fmt.Println("Exported Accounts:")
	fmt.Println("==================")
	for i, account := range accounts {
		fmt.Printf("[%d] DID: %s\n", i, account.DID)
		fmt.Printf("    Balance: %.2f RBT\n", account.Balance)
		fmt.Printf("    Pledged: %.2f | Locked: %.2f | Pinned: %.2f\n\n",
			account.PledgedRBT, account.LockedRBT, account.PinnedRBT)
	}

	fmt.Println("Usage:")
	fmt.Printf("  ./break-nlss transfer --from-file %s --sender-index 0 --receiver <DID> --amount <AMOUNT>\n", *output)
}

func runGenerateKey() {
	genCmd := flag.NewFlagSet("generate-key", flag.ExitOnError)

	outputDir := genCmd.String("output", "./preset", "Output directory for key files")

	genCmd.Parse(os.Args[2:])

	fmt.Printf("Generating new EC key pair (P-256)...\n")
	fmt.Printf("Output directory: %s\n\n", *outputDir)

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	privateKeyPath := fmt.Sprintf("%s/privatekey.pem", *outputDir)
	publicKeyPath := fmt.Sprintf("%s/publickey.pem", *outputDir)

	// Generate keys
	_, err := rubix.GenerateAndSaveKeys(privateKeyPath, publicKeyPath)
	if err != nil {
		fmt.Printf("Error generating keys: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Key pair generated successfully!\n")
	fmt.Printf("  Private key: %s\n", privateKeyPath)
	fmt.Printf("  Public key: %s\n", publicKeyPath)
	fmt.Println("\nIMPORTANT: Keep your private key secure and never share it!")
}

func runBreakNLSS() {
	breakCmd := flag.NewFlagSet("break-nlss", flag.ExitOnError)

	didInput := breakCmd.String("did", "", "DID string or path to file containing DIDs (required)")

	breakCmd.Parse(os.Args[2:])

	// Validate required flags
	if *didInput == "" {
		fmt.Println("Error: --did is required")
		fmt.Println("\nUsage:")
		fmt.Println("  Single DID:")
		fmt.Println("    break-nlss break-nlss --did bafybmifeh7csi6wuuwqd3c7cxcwk5k3e3nd2f73x2hxa2teojhkd6ztdse")
		fmt.Println("  Multiple DIDs from file:")
		fmt.Println("    break-nlss break-nlss --did dids.txt")
		breakCmd.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Determine if input is a file or a single DID
	var dids []string
	if _, err := os.Stat(*didInput); err == nil {
		// It's a file, read DIDs from it
		fmt.Printf("Reading DIDs from file: %s\n", *didInput)
		dids, err = readDIDsFromFile(*didInput)
		if err != nil {
			fmt.Printf("Error reading DIDs from file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d DIDs to process\n\n", len(dids))
	} else {
		// It's a single DID
		dids = []string{*didInput}
	}

	// Process each DID
	successCount := 0
	failCount := 0

	for i, did := range dids {
		did = strings.TrimSpace(did)
		if did == "" {
			continue // Skip empty lines
		}

		fmt.Printf("[%d/%d] Processing DID: %s\n", i+1, len(dids), did)
		fmt.Println("============================================")

		// Get input image paths from config
		didImagePath, pubSharePath, err := cfg.GetNLSSImagePaths(did)
		if err != nil {
			fmt.Printf("❌ Error constructing paths: %v\n\n", err)
			failCount++
			continue
		}

		// Get output path from config
		outputPath, err := cfg.GetNLSSOutputPath(did)
		if err != nil {
			fmt.Printf("❌ Error constructing output path: %v\n\n", err)
			failCount++
			continue
		}

		fmt.Printf("  DID Image: %s\n", didImagePath)
		fmt.Printf("  Public Share: %s\n", pubSharePath)
		fmt.Printf("  Output: %s\n", outputPath)

		// Check if input files exist
		if _, err := os.Stat(didImagePath); os.IsNotExist(err) {
			fmt.Printf("❌ Error: DID image file not found: %s\n\n", didImagePath)
			failCount++
			continue
		}

		if _, err := os.Stat(pubSharePath); os.IsNotExist(err) {
			fmt.Printf("❌ Error: Public share file not found: %s\n\n", pubSharePath)
			failCount++
			continue
		}

		// Run the BreakNLSS algorithm
		err = nlss.BreakNLSSFromFiles(didImagePath, pubSharePath, outputPath)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			failCount++
			continue
		}

		fmt.Printf("✓ Successfully reconstructed private share!\n")
		fmt.Printf("  Saved to: %s\n\n", outputPath)
		successCount++
	}

	// Print summary
	fmt.Println("============================================")
	fmt.Println("Summary:")
	fmt.Printf("  Total DIDs: %d\n", len(dids))
	fmt.Printf("  Successful: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failCount)
	fmt.Println("\nIMPORTANT: Keep your private shares secure and never share them!")
}

// readDIDsFromFile reads DIDs from a text file (one per line)
func readDIDsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var dids []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			dids = append(dids, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return dids, nil
}
