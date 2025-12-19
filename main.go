package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/hades/podx/crypto"
	"github.com/hades/podx/keygen"
	"github.com/hades/podx/parser"
	"github.com/hades/podx/project"
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
		// Launch TUI mode when no arguments
		if err := tui.Run(Version); err != nil {
			fmt.Println("Error running TUI:", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "tui":
		// Explicit TUI launch
		if err := tui.Run(Version); err != nil {
			fmt.Println("Error running TUI:", err)
			os.Exit(1)
		}
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
	case "update":
		handleUpdate()
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

OTHER:
  tui        Launch interactive TUI mode
  update     Self-update to latest version
  version    Show version info

USAGE:
  podx                                   # Launch TUI mode
  podx tui                               # Launch TUI mode
  podx init                              # Init project
  podx add-recipient -n "Name" -k KEY    # Add team member
  podx encrypt-all                       # Encrypt all secrets
  podx decrypt-all                       # Decrypt all secrets
  podx keygen -t age                     # Generate Age key
  podx update                            # Update to latest
  podx encrypt -a aes-gcm -i F -o F.enc  # Encrypt file
  podx env encrypt -i .env -o .env.podx  # Encrypt .env`)
}

func handleEncrypt(args []string) {
	fs := flag.NewFlagSet("encrypt", flag.ExitOnError)
	algo := fs.String("a", "aes-gcm", "Algorithm (aes-gcm or chacha20)")
	fs.String("algorithm", "aes-gcm", "")
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

	if *input == "" {
		fmt.Println("Error: input (-i) is required")
		os.Exit(1)
	}
	if *output == "" {
		*output = deriveEncryptOutput(*input)
	}
	if *output == "" {
		fmt.Println("Error: output (-o) is required")
		os.Exit(1)
	}
	if samePath(*input, *output) {
		fmt.Println("Error: output must be different from input")
		os.Exit(1)
	}

	// Get password
	pass := getPassword(*password, "Enter password: ")

	// Derive key
	key, salt, err := crypto.DeriveKey([]byte(pass), nil)
	if err != nil {
		fmt.Println("Error deriving key:", err)
		os.Exit(1)
	}

	// Read input file
	plaintext, err := os.ReadFile(*input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}

	// Get encryptor
	enc, err := crypto.NewEncryptor(crypto.Algorithm(*algo))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Encrypt
	ciphertext, err := enc.Encrypt(plaintext, key)
	if err != nil {
		fmt.Println("Error encrypting:", err)
		os.Exit(1)
	}

	// Write output: [salt (16 bytes)][algo (1 byte)][ciphertext]
	algoB := byte(0)
	if *algo == "chacha20" {
		algoB = 1
	}

	out := append(salt, algoB)
	out = append(out, ciphertext...)

	if err := os.WriteFile(*output, out, 0600); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Encrypted %s → %s (algorithm: %s)\n", *input, *output, *algo)
	if err := removeOriginal(*input, *output); err != nil {
		fmt.Println("Error deleting original:", err)
		os.Exit(1)
	}
	fmt.Printf("🗑️  Deleted original: %s\n", *input)
}

func handleDecrypt(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	fs.String("a", "", "Algorithm (auto-detected)")
	fs.String("algorithm", "", "")
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

	if *input == "" {
		fmt.Println("Error: input (-i) is required")
		os.Exit(1)
	}
	if *output == "" {
		*output = deriveDecryptOutput(*input)
	}
	if *output == "" {
		fmt.Println("Error: output (-o) is required")
		os.Exit(1)
	}
	if samePath(*input, *output) {
		fmt.Println("Error: output must be different from input")
		os.Exit(1)
	}

	// Get password
	pass := getPassword(*password, "Enter password: ")

	// Read input file
	data, err := os.ReadFile(*input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}

	if len(data) < crypto.SaltSize+1 {
		fmt.Println("Error: file too small or corrupted")
		os.Exit(1)
	}

	// Extract salt and algo
	salt := data[:crypto.SaltSize]
	algoB := data[crypto.SaltSize]
	ciphertext := data[crypto.SaltSize+1:]

	algo := crypto.AlgoAESGCM
	if algoB == 1 {
		algo = crypto.AlgoChaCha20
	}

	// Derive key
	key, err := crypto.DeriveKeyWithSalt([]byte(pass), salt)
	if err != nil {
		fmt.Println("Error deriving key:", err)
		os.Exit(1)
	}

	// Get encryptor
	enc, err := crypto.NewEncryptor(algo)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Decrypt
	plaintext, err := enc.Decrypt(ciphertext, key)
	if err != nil {
		fmt.Println("Error decrypting (wrong password?):", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, plaintext, 0600); err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Decrypted %s → %s (algorithm: %s)\n", *input, *output, algo)
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

	if *input == "" {
		fmt.Println("Error: input (-i) is required")
		os.Exit(1)
	}

	switch subcmd {
	case "encrypt":
		if *output == "" {
			*output = deriveEnvEncryptOutput(*input)
		}
		if *output == "" {
			fmt.Println("Error: output (-o) is required")
			os.Exit(1)
		}
		if samePath(*input, *output) {
			fmt.Println("Error: output must be different from input")
			os.Exit(1)
		}
		handleEnvEncrypt(*input, *output, *algo, *password)
	case "decrypt":
		if *output == "" {
			*output = deriveDecryptOutput(*input)
		}
		if *output == "" {
			fmt.Println("Error: output (-o) is required")
			os.Exit(1)
		}
		if samePath(*input, *output) {
			fmt.Println("Error: output must be different from input")
			os.Exit(1)
		}
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
	if err := removeOriginal(input, output); err != nil {
		fmt.Println("Error deleting original:", err)
		os.Exit(1)
	}
	fmt.Printf("🗑️  Deleted original: %s\n", input)
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

func samePath(a, b string) bool {
	aAbs, errA := filepath.Abs(a)
	bAbs, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return aAbs == bAbs
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func removeOriginal(input, output string) error {
	if samePath(input, output) {
		return fmt.Errorf("output must be different from input")
	}
	return os.Remove(input)
}

func deriveEncryptOutput(input string) string {
	if input == "" {
		return ""
	}
	return input + ".enc"
}

func deriveEnvEncryptOutput(input string) string {
	if input == "" {
		return ""
	}
	return input + ".podx"
}

func deriveDecryptOutput(input string) string {
	if input == "" {
		return ""
	}
	switch {
	case strings.HasSuffix(input, ".enc"):
		return strings.TrimSuffix(input, ".enc")
	case strings.HasSuffix(input, ".podx"):
		return strings.TrimSuffix(input, ".podx")
	case strings.HasSuffix(input, ".age"):
		return strings.TrimSuffix(input, ".age")
	default:
		return input + ".dec"
	}
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

	// Check for updates in background
	if newVersion, available := updater.CheckUpdate(Version); available {
		fmt.Printf("\n📦 New version available: %s\n", newVersion)
		fmt.Println("   Run 'podx update' to upgrade")
	}
}

func handleUpdate() {
	if err := updater.Update(Version); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
