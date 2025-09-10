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
	saltStr := flag.String("salt", "", "Salt string (non-optional)")

	par := flag.Uint("p", 12, "Parallelism (threads)")
	mem := flag.Uint("mem", 4096, "Memory cost in MiB")
	tim := flag.Uint("time", 2, "Time cost (iterations)")
	keyLen := flag.Uint("len", 128, "Derived key length in bytes")
	of := flag.String("out", "", "Path to save raw binary key to")
	useHex := flag.Bool("hex", false, "Use hex, only for text output")
	passStdin := flag.Bool("stdin", false, "Read password from stdin")
	flag.Parse()

	// For disabling timestamps in errors
	log.SetFlags(0)

	if *saltStr == "" {
		log.Fatal("Error: Provide a salt via the -salt flag.")
	}

	if *useHex && *of != "" {
		log.Fatal("Error: -hex cannot be used alongside -out.")
	}

	if len(*saltStr) < 16 {
		fmt.Fprintln(os.Stderr, "Warning: a salt of at least 16",
		"characters is recommended for maximum security.")
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
	dk := argon2.IDKey(
		passwordBytes,
		[]byte(*saltStr),
		uint32(*tim),
		uint32(*mem*1024),
		uint8(*par),
		uint32(*keyLen),
	) // derived key

	// We clear the array after we're done, so nobody can read it
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	// Output time
	if *of != "" {
		err := os.WriteFile(*of, dk, 0600)
		if err != nil {
			log.Fatalf("Error writing key file to %s: %v", *of, err)
		}
		fmt.Fprintf(os.Stderr, "Success! Raw key saved to %s\n", *of)
	} else {
		var ek string // Encoded key
		if *useHex {
			ek = hex.EncodeToString(dk)
		} else {
			ek = base64.StdEncoding.EncodeToString(dk)
		}
		fmt.Println(ek)
	}
}
