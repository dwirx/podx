package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
	"github.com/hades/podx/parser"
	"github.com/hades/podx/project"
	"github.com/hades/podx/security"
	"github.com/hades/podx/tui"
	"github.com/hades/podx/updater"
	"golang.org/x/term"
)

// Version info - injected at build time via ldflags
var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

const banner = `
╔═══════════════════════════════════════════╗
║  ██████╗  ██████╗ ██████╗ ██╗  ██╗        ║
║  ██╔══██╗██╔═══██╗██╔══██╗╚██╗██╔╝        ║
║  ██████╔╝██║   ██║██║  ██║ ╚███╔╝         ║
║  ██╔═══╝ ██║   ██║██║  ██║ ██╔██╗         ║
║  ██║     ╚██████╔╝██████╔╝██╔╝ ██╗        ║
║  ╚═╝      ╚═════╝ ╚═════╝ ╚═╝  ╚═╝        ║
║  🔐 Encryption Tool %s                    ║
╚═══════════════════════════════════════════╝
`

func main() {
	if len(os.Args) < 2 {
		// Launch TUI when no arguments
		if err := tui.Run(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		handleInit()
	case "add-recipient":
		handleAddRecipient(os.Args[2:])
	case "encrypt-all":
		handleEncryptAll()
	case "decrypt-all":
		handleDecryptAll()
	case "status":
		handleStatus()
	case "check":
		handleCheck(os.Args[2:])
	case "hook":
		if len(os.Args) < 3 {
			fmt.Println("Usage: podx hook <install|uninstall|status>")
			os.Exit(1)
		}
		handleHook(os.Args[2], os.Args[3:])
	case "encrypt":
		handleEncrypt(os.Args[2:])
	case "decrypt":
		handleDecrypt(os.Args[2:])
	case "env":
		if len(os.Args) < 3 {
			fmt.Println("Usage: podx env <encrypt|decrypt> [options]")
			os.Exit(1)
		}
		handleEnv(os.Args[2], os.Args[3:])
	case "keygen":
		handleKeygen(os.Args[2:])
	case "key-info":
		handleKeyInfo()
	case "update":
		handleUpdate(os.Args[2:])
	case "rollback":
		handleRollback()
	case "shamir":
		if len(os.Args) < 3 {
			fmt.Println("Usage: podx shamir <split|combine> [options]")
			os.Exit(1)
		}
		handleShamir(os.Args[2], os.Args[3:])
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(banner, Version)
	fmt.Println(`
PROJECT COMMANDS:
  init           Initialize PODX project (.podx.yaml)
  add-recipient  Add team member to project
  encrypt-all    Encrypt all secrets in project
  decrypt-all    Decrypt all secrets in project
  status         Show project status

FILE COMMANDS:
  encrypt    Encrypt a single file
  decrypt    Decrypt a single file
  env        Encrypt/decrypt .env file (format-preserving)
  keygen     Generate Age or GPG key pair
  key-info   Show your Age public key

OTHER:
  check      Check for unencrypted secrets
  hook       Manage pre-commit hook
  shamir     Split/combine secrets with Shamir Secret Sharing
  update     Self-update to latest version (--beta for beta)
  rollback   Rollback to previous version after update
  version    Show version info

USAGE:
  podx init                              # Init project
  podx add-recipient -n "Name" -k KEY    # Add team member
  podx encrypt-all                       # Encrypt all secrets
  podx decrypt-all                       # Decrypt all secrets
  podx keygen -t age                     # Generate Age key
  podx update                            # Update to latest
  podx encrypt -a aes-gcm -i F -o F.enc  # Encrypt file
  podx env encrypt -i .env -o .env.podx  # Encrypt .env
  podx shamir split -i key.txt -t 3 -n 5 # Split key to 5 shares (need 3)
  podx shamir combine -d ./shares        # Combine shares`)
}

func handleEncrypt(args []string) {
	fs := flag.NewFlagSet("encrypt", flag.ExitOnError)
	algo := fs.String("a", "", "Algorithm (aes-gcm, xchacha20, cascade)")
	fs.String("algorithm", "", "")
	mode := fs.String("m", "normal", "Encryption mode (normal or paranoid)")
	fs.String("mode", "normal", "")
	cipher := fs.String("c", "", "Cipher (aes-gcm, xchacha20) - overrides mode default")
	fs.String("cipher", "", "")
	input := fs.String("i", "", "Input file")
	fs.String("input", "", "")
	output := fs.String("o", "", "Output file")
	fs.String("output", "", "")
	password := fs.String("p", "", "Password")
	fs.String("password", "", "")
	memoryMB := fs.Uint("memory", 0, "Argon2id memory in MB (0 for auto)")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: input (-i) and output (-o) are required")
		os.Exit(1)
	}

	// Parse mode
	encMode, err := crypto.ParseEncryptMode(*mode)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Determine cipher type
	var cipherType crypto.CipherType
	if *cipher != "" {
		cipherType, err = crypto.ParseCipher(*cipher)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else if *algo != "" {
		// Legacy -a flag support
		cipherType, err = crypto.ParseCipher(*algo)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else {
		// Default based on mode
		if encMode == crypto.ModeParanoid {
			cipherType = crypto.CipherCascade
		} else {
			cipherType = crypto.CipherAESGCM
		}
	}

	// Get password
	pass := getPassword(*password, "Enter password: ")

	// Build encryption options
	opts := &crypto.EncryptOptions{
		Mode:     encMode,
		Cipher:   cipherType,
		MemoryMB: uint32(*memoryMB),
	}

	// Check if streaming should be used
	useStreaming, _ := crypto.ShouldUseStreaming(*input)
	if useStreaming {
		fmt.Println("Large file detected, using streaming encryption...")
		err = crypto.StreamingEncryptFile(*input, *output, []byte(pass), opts, func(processed, total int64) {
			pct := float64(processed) / float64(total) * 100
			fmt.Printf("\rProgress: %.1f%%", pct)
		})
		fmt.Println()
	} else {
		// Read input file
		plaintext, readErr := os.ReadFile(*input)
		if readErr != nil {
			fmt.Println("Error reading input:", readErr)
			os.Exit(1)
		}

		// Encrypt using v2 format
		ciphertext, encErr := crypto.EncryptV2(plaintext, []byte(pass), opts)
		if encErr != nil {
			fmt.Println("Error encrypting:", encErr)
			os.Exit(1)
		}

		err = os.WriteFile(*output, ciphertext, 0600)
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	cipherName := crypto.CipherString(cipherType)
	modeName := crypto.EncryptModeString(encMode)
	fmt.Printf("✓ Encrypted %s → %s (mode: %s, cipher: %s)\n", *input, *output, modeName, cipherName)
}

func handleDecrypt(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	input := fs.String("i", "", "Input file")
	fs.String("input", "", "")
	output := fs.String("o", "", "Output file")
	fs.String("output", "", "")
	password := fs.String("p", "", "Password")
	fs.String("password", "", "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: input (-i) and output (-o) are required")
		os.Exit(1)
	}

	// Get password
	pass := getPassword(*password, "Enter password: ")

	// Use auto-detection for decryption
	err := crypto.DecryptFileAuto(*input, *output, []byte(pass), func(processed, total int64) {
		pct := float64(processed) / float64(total) * 100
		fmt.Printf("\rProgress: %.1f%%", pct)
	})

	if err != nil {
		fmt.Println("Error decrypting:", err)
		os.Exit(1)
	}

	// Get format info for display
	data, _ := os.ReadFile(*input)
	info, _ := crypto.EncryptV2Info(data)
	fmt.Printf("✓ Decrypted %s → %s (%s)\n", *input, *output, info)
}

func handleEnv(subcmd string, args []string) {
	fs := flag.NewFlagSet("env", flag.ExitOnError)
	algo := fs.String("a", "aes-gcm", "Algorithm (aes-gcm or chacha20)")
	fs.String("algorithm", "aes-gcm", "")
	input := fs.String("i", "", "Input .env file")
	fs.String("input", "", "")
	output := fs.String("o", "", "Output .env file")
	fs.String("output", "", "")
	password := fs.String("p", "", "Password")
	fs.String("password", "", "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *input == "" || *output == "" {
		fmt.Println("Error: input (-i) and output (-o) are required")
		os.Exit(1)
	}

	switch subcmd {
	case "encrypt":
		handleEnvEncrypt(*input, *output, *algo, *password)
	case "decrypt":
		handleEnvDecrypt(*input, *output, *password)
	default:
		fmt.Printf("Unknown env subcommand: %s\n", subcmd)
		os.Exit(1)
	}
}

func handleEnvEncrypt(input, output, algo, password string) {
	pass := getPassword(password, "Enter password: ")

	// Parse .env
	entries, err := parser.ParseEnvFile(input)
	if err != nil {
		fmt.Println("Error parsing .env:", err)
		os.Exit(1)
	}

	// Derive key
	key, salt, err := crypto.DeriveKey([]byte(pass), nil)
	if err != nil {
		fmt.Println("Error deriving key:", err)
		os.Exit(1)
	}

	// Encrypt values
	if err := parser.EncryptEnvValues(entries, key, crypto.Algorithm(algo)); err != nil {
		fmt.Println("Error encrypting:", err)
		os.Exit(1)
	}

	// Add salt sebagai comment di awal file
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	saltEntry := parser.EnvEntry{
		IsComment: true,
		Comment:   fmt.Sprintf("# IRONVAULT_SALT=%s", saltB64),
	}
	entries = append([]parser.EnvEntry{saltEntry}, entries...)

	// Write output
	if err := parser.WriteEnvFile(output, entries); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Encrypted .env: %s → %s (algorithm: %s)\n", input, output, algo)
}

func handleEnvDecrypt(input, output, password string) {
	pass := getPassword(password, "Enter password: ")

	// Parse .env
	entries, err := parser.ParseEnvFile(input)
	if err != nil {
		fmt.Println("Error parsing .env:", err)
		os.Exit(1)
	}

	// Extract salt from comment
	var salt []byte
	var cleanEntries []parser.EnvEntry

	for _, entry := range entries {
		if entry.IsComment && strings.HasPrefix(entry.Comment, "# IRONVAULT_SALT=") {
			saltB64 := strings.TrimPrefix(entry.Comment, "# IRONVAULT_SALT=")
			salt, err = base64.StdEncoding.DecodeString(saltB64)
			if err != nil {
				fmt.Println("Error decoding salt:", err)
				os.Exit(1)
			}
		} else {
			cleanEntries = append(cleanEntries, entry)
		}
	}

	if salt == nil {
		fmt.Println("Error: no salt found in encrypted file")
		os.Exit(1)
	}

	// Derive key
	key, err := crypto.DeriveKeyWithSalt([]byte(pass), salt)
	if err != nil {
		fmt.Println("Error deriving key:", err)
		os.Exit(1)
	}

	// Decrypt values
	if err := parser.DecryptEnvValues(cleanEntries, key); err != nil {
		fmt.Println("Error decrypting:", err)
		os.Exit(1)
	}

	// Write output
	if err := parser.WriteEnvFile(output, cleanEntries); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Decrypted .env: %s → %s\n", input, output)
}

func getPassword(provided, prompt string) string {
	if provided != "" {
		return provided
	}

	fmt.Print(prompt)

	// Coba baca dari terminal
	if term.IsTerminal(int(syscall.Stdin)) {
		password, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading password:", err)
			os.Exit(1)
		}
		return string(password)
	}

	// Fallback untuk non-terminal (piped input)
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Println("Error reading password:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(password)
}

func handleKeygen(args []string) {
	fs := flag.NewFlagSet("keygen", flag.ExitOnError)
	keyType := fs.String("t", "age", "Key type (age or gpg)")
	fs.String("type", "age", "")
	name := fs.String("n", "", "Name for GPG key")
	fs.String("name", "", "")
	email := fs.String("e", "", "Email for GPG key")
	fs.String("email", "", "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	switch *keyType {
	case "age":
		result, err := keygen.GenerateAge()
		if err != nil {
			fmt.Println("Error generating Age key:", err)
			os.Exit(1)
		}
		keygen.PrintKeygenResult(result)

	case "gpg":
		if *name == "" || *email == "" {
			fmt.Println("Error: name (-n) and email (-e) are required for GPG key generation")
			os.Exit(1)
		}
		result, err := keygen.GenerateGPG(*name, *email)
		if err != nil {
			fmt.Println("Error generating GPG key:", err)
			os.Exit(1)
		}
		keygen.PrintKeygenResult(result)

	default:
		fmt.Printf("Unknown key type: %s (supported: age, gpg)\n", *keyType)
		os.Exit(1)
	}
}

func handleKeyInfo() {
	info := keygen.GetAgeKeyInfo()

	width := 70
	fmt.Printf("╔%s╗\n", strings.Repeat("═", width))
	fmt.Printf("║%s║\n", centerText("🔑 Your Age Key Information", width))
	fmt.Printf("╠%s╣\n", strings.Repeat("═", width))

	if info.HasKey {
		printBoxRow("Status:", "✓ Key found", width)
		printBoxRow("Key file:", info.KeyFilePath, width)
		fmt.Printf("╠%s╣\n", strings.Repeat("═", width))
		printBoxRow("Public Key (share with team):", "", width)
		printBoxRow("  "+info.PublicKey, "", width)
	} else {
		printBoxRow("Status:", "✗ No key found", width)
		printBoxRow("", "", width)
		printBoxRow("Generate a new key with:", "podx keygen -t age", width)
	}

	fmt.Printf("╚%s╝\n", strings.Repeat("═", width))
}

func centerText(text string, width int) string {
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", padding), text, strings.Repeat(" ", width-padding-len(text)))
}

func printBoxRow(label, value string, width int) {
	content := label
	if value != "" {
		content = label + " " + value
	}
	// Truncate if too long
	if len(content) > width-2 {
		content = content[:width-5] + "..."
	}
	fmt.Printf("║ %s%s║\n", content, strings.Repeat(" ", width-len(content)-1))
}

// Project commands

func handleInit() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}

	p, err := project.Init(cwd)
	if err != nil {
		fmt.Println("Error initializing project:", err)
		os.Exit(1)
	}

	project.PrintInitSuccess(p)
}

func handleAddRecipient(args []string) {
	fs := flag.NewFlagSet("add-recipient", flag.ExitOnError)
	name := fs.String("n", "", "Recipient name")
	fs.String("name", "", "")
	key := fs.String("k", "", "Recipient Age public key")
	fs.String("key", "", "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *name == "" || *key == "" {
		fmt.Println("Error: name (-n) and key (-k) are required")
		fmt.Println("Usage: podx add-recipient -n 'Team Member' -k age1xxx...")
		os.Exit(1)
	}

	cwd, _ := os.Getwd()
	p, err := project.Load(cwd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := p.AddRecipient(*name, *key); err != nil {
		fmt.Println("Error adding recipient:", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Added recipient: %s (%s...)\n", *name, (*key)[:20])
}

func handleEncryptAll() {
	cwd, _ := os.Getwd()
	p, err := project.Load(cwd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	count, err := p.EncryptAll()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if count == 0 {
		fmt.Println("No files to encrypt")
	} else {
		fmt.Printf("\n🔐 Encrypted %d file(s)\n", count)
	}
}

func handleDecryptAll() {
	cwd, _ := os.Getwd()
	p, err := project.Load(cwd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	count, err := p.DecryptAll()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if count == 0 {
		fmt.Println("No files to decrypt")
	} else {
		fmt.Printf("\n🔓 Decrypted %d file(s)\n", count)
	}
}

func handleStatus() {
	cwd, _ := os.Getwd()
	p, err := project.Load(cwd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println(p.Status())
}

func printVersion() {
	fmt.Printf("PODX %s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Check for updates
	info := updater.CheckUpdate(Version)
	if info.Available {
		fmt.Printf("\n📦 New version available: %s → %s\n", info.CurrentVersion, info.LatestVersion)
		if info.DownloadSize > 0 {
			fmt.Printf("   Size: %s\n", updater.FormatSize(info.DownloadSize))
		}
		fmt.Println("   Run 'podx update' to upgrade")
	}
}

func handleUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	beta := fs.Bool("beta", false, "Update to beta version")
	fs.Parse(args)

	if err := updater.Update(Version, *beta); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func handleRollback() {
	if err := updater.Rollback(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func handleCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	preCommit := fs.Bool("pre-commit", false, "Pre-commit mode (exit code only)")
	fix := fs.Bool("fix", false, "Auto-fix gitignore issues")
	fs.Parse(args)

	cwd, _ := os.Getwd()
	result := security.CheckProject(cwd, *fix)

	output := security.FormatResult(result, *preCommit)
	if output != "" {
		fmt.Print(output)
	}

	if !result.Passed {
		os.Exit(1)
	}
}

func handleHook(subcmd string, args []string) {
	cwd, _ := os.Getwd()

	switch subcmd {
	case "install":
		if err := security.InstallHook(cwd); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Pre-commit hook installed")
		fmt.Println("\nThe hook will run 'podx check' before each commit.")
		fmt.Println("If unencrypted secrets are found, the commit will be blocked.")
		fmt.Println("\nTo uninstall: podx hook uninstall")

	case "uninstall":
		if err := security.UninstallHook(cwd); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Pre-commit hook removed")

	case "status":
		if security.IsHookInstalled(cwd) {
			fmt.Println("PODX pre-commit hook is installed")
		} else {
			fmt.Println("PODX pre-commit hook is not installed")
			fmt.Println("Run 'podx hook install' to enable")
		}

	default:
		fmt.Printf("Unknown hook command: %s\n", subcmd)
		fmt.Println("Usage: podx hook <install|uninstall|status>")
		os.Exit(1)
	}
}

func handleShamir(subcmd string, args []string) {
	switch subcmd {
	case "split":
		handleShamirSplit(args)
	case "combine":
		handleShamirCombine(args)
	case "presets":
		handleShamirPresets()
	default:
		fmt.Printf("Unknown shamir command: %s\n", subcmd)
		fmt.Println("Usage: podx shamir <split|combine|presets> [options]")
		os.Exit(1)
	}
}

func handleShamirSplit(args []string) {
	fs := flag.NewFlagSet("shamir split", flag.ExitOnError)
	input := fs.String("i", "", "Input file containing secret")
	fs.String("input", "", "")
	output := fs.String("o", "", "Output directory for shares (default: ./shares)")
	fs.String("output", "", "")
	threshold := fs.Int("t", 0, "Threshold (minimum shares needed)")
	fs.Int("threshold", 0, "")
	total := fs.Int("n", 0, "Total number of shares")
	fs.Int("total", 0, "")
	preset := fs.String("p", "", "Use preset (2-of-3, 3-of-5, 4-of-7, 5-of-9)")
	fs.String("preset", "", "")
	qrcode := fs.Bool("qr", false, "Generate QR codes for shares")
	fs.Bool("qrcode", false, "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *input == "" {
		fmt.Println("Error: input file (-i) is required")
		os.Exit(1)
	}

	// Determine threshold and total
	var t, n int
	if *preset != "" {
		p, err := crypto.GetPreset(*preset)
		if err != nil {
			fmt.Println("Error:", err)
			fmt.Println("Available presets: 2-of-3, 3-of-5, 4-of-7, 5-of-9")
			os.Exit(1)
		}
		t = p.Threshold
		n = p.TotalShares
	} else if *threshold > 0 && *total > 0 {
		t = *threshold
		n = *total
	} else {
		fmt.Println("Error: specify -p (preset) or both -t (threshold) and -n (total)")
		fmt.Println("Examples:")
		fmt.Println("  podx shamir split -i secret.key -p 3-of-5")
		fmt.Println("  podx shamir split -i secret.key -t 3 -n 5")
		os.Exit(1)
	}

	// Read secret
	secret, err := os.ReadFile(*input)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Split
	shares, err := crypto.ShamirSplit(secret, t, n)
	if err != nil {
		fmt.Printf("Error splitting secret: %v\n", err)
		os.Exit(1)
	}

	// Determine output directory
	outDir := *output
	if outDir == "" {
		outDir = "./shares"
	}

	// Save shares
	paths, err := crypto.SaveAllShares(shares, outDir)
	if err != nil {
		fmt.Printf("Error saving shares: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Secret split into %d shares (threshold: %d)\n", n, t)
	fmt.Println("  Shares saved to:", outDir)
	for _, path := range paths {
		fmt.Println("   ", path)
	}

	// Generate QR codes if requested
	if *qrcode {
		qrPaths, err := crypto.SaveAllShareQRCodes(shares, outDir, nil)
		if err != nil {
			fmt.Printf("Warning: Failed to generate QR codes: %v\n", err)
		} else {
			fmt.Println("\n  QR codes generated:")
			for _, path := range qrPaths {
				fmt.Println("   ", path)
			}
		}
	}

	fmt.Printf("\n⚠️  IMPORTANT: Distribute shares to different locations/people.\n")
	fmt.Printf("   At least %d shares are needed to recover the secret.\n", t)
}

func handleShamirCombine(args []string) {
	fs := flag.NewFlagSet("shamir combine", flag.ExitOnError)
	dir := fs.String("d", "", "Directory containing share files")
	fs.String("dir", "", "")
	output := fs.String("o", "", "Output file for recovered secret")
	fs.String("output", "", "")
	files := fs.String("f", "", "Comma-separated list of share files")
	fs.String("files", "", "")

	if err := fs.Parse(args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if *output == "" {
		fmt.Println("Error: output file (-o) is required")
		os.Exit(1)
	}

	var shares []*crypto.Share
	var err error

	if *dir != "" {
		// Load from directory
		shares, err = crypto.LoadSharesFromDir(*dir)
		if err != nil {
			fmt.Printf("Error loading shares: %v\n", err)
			os.Exit(1)
		}
	} else if *files != "" {
		// Load from specific files
		fileList := strings.Split(*files, ",")
		for _, f := range fileList {
			f = strings.TrimSpace(f)
			share, loadErr := crypto.LoadShare(f)
			if loadErr != nil {
				fmt.Printf("Error loading share %s: %v\n", f, loadErr)
				os.Exit(1)
			}
			shares = append(shares, share)
		}
	} else {
		fmt.Println("Error: specify -d (directory) or -f (files)")
		os.Exit(1)
	}

	if len(shares) == 0 {
		fmt.Println("Error: no shares found")
		os.Exit(1)
	}

	// Validate shares
	if err := crypto.ValidateShares(shares); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d shares:\n", len(shares))
	for _, s := range shares {
		fmt.Printf("  - %s\n", crypto.ShareInfo(s))
	}

	// Combine
	secret, err := crypto.ShamirCombine(shares)
	if err != nil {
		fmt.Printf("Error combining shares: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(*output, secret, 0600); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Secret recovered and saved to: %s\n", *output)
}

func handleShamirPresets() {
	fmt.Println("Available Shamir Secret Sharing presets:\n")
	for _, p := range crypto.ShamirPresets {
		fmt.Printf("  %-8s  %s\n", p.Name, p.Description)
	}
	fmt.Println("\nUsage: podx shamir split -i <file> -p <preset>")
}
