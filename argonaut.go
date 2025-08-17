package main

import ( 
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

func main() {
	// Command line flags
	par := flag.Uint("p", 4, "Parallelism (threads). Set this to the number of threads you have.")
	mem := flag.Uint("mem", 1024, "Memory cost in MiB. Set this as high as comfortably possible for your system.")
	tim := flag.Uint("time", 1, "Time cost (iterations). If you're unhappy with the memory you have, increase this to increase the cost for brute force attacks.")
	saltStr := flag.String("salt", "", "Salt string (required).")
	keyLen := flag.Uint("len", 128, "Derived key length in bytes.")
	useHex := flag.Bool("hex", false, "Use hex, only for text output.")
	outFile := flag.String("out", "", "Path to save the raw binary key to.")
	passStdin := flag.Bool("stdin", false, "Read master password from stdin.")
	flag.Parse()

	if *saltStr == "" {
		log.Fatal("Error: a salt must be provided via the -salt flag.")
	}

	if *useHex && *outFile != "" {
		log.Fatal("Error: -hex cannot be used alongside -out.")
	}

	if len(*saltStr) < 16 {
		fmt.Fprintln(os.Stderr, "Warning: a salt of at least 16 characters is recommended for maximum security.")
	}


	var passwordBytes []byte
	var err error

	if *passStdin {
		passwordBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading password from stdin: %v", err)
		}
	} else {
		fmt.Fprint(os.Stderr, "Enter Master Password: ")
		passwordBytes, err = term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatalf("Error reading password: %v", err)
		}
		fmt.Fprintln(os.Stderr)
	}

	// Where the magic happens
	derivedKey := argon2.IDKey(
		passwordBytes,
		[]byte(*saltStr),
		uint32(*tim),
		uint32(*mem*1024),
		uint8(*par),
		uint32(*keyLen),
	)

	// We clear the array after we're done, so nobody can read it
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	// Output time
	if *outFile != "" {
		err := os.WriteFile(*outFile, derivedKey, 0600)
		if err != nil {
			log.Fatalf("Error writing key file to %s: %v", *outFile, err)
		}
		fmt.Fprintf(os.Stderr, "Success! Raw key saved to %s\n", *outFile)
	} else {
		var encodedKey string
		if *useHex {
			encodedKey = hex.EncodeToString(derivedKey)
		} else {
			encodedKey = base64.StdEncoding.EncodeToString(derivedKey)
		}
		fmt.Println(encodedKey)
	}
}
