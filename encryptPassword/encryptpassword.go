package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// Takes two strings, cryptoText and keyString.
// cryptoText is the text to be decrypted and the keyString is the key to use for the decryption.
// The function will output the resulting plain text string with an error variable.
func decryptString(cryptoText string, keyString string) (plainTextString string, err error) {

	// Format the keyString so that it's 32 bytes.
	newKeyString, err := hashTo32Bytes(keyString)

	// Encode the cryptoText to base 64.
	cipherText, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher([]byte(newKeyString))

	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

// Takes two string, plainText and keyString.
// plainText is the text that needs to be encrypted by keyString.
// The function will output the resulting crypto text and an error variable.
func encryptString(plainText string, keyString string) (cipherTextString string, err error) {

	// Format the keyString so that it's 32 bytes.
	newKeyString, err := hashTo32Bytes(keyString)

	if err != nil {
		return "", err
	}

	key := []byte(newKeyString)
	value := []byte(plainText)

	block, err := aes.NewCipher(key)

	if err != nil {
		panic(err)
	}

	cipherText := make([]byte, aes.BlockSize+len(value))

	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], value)

	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// As we cannot use a variable length key, we must cut the users key
// up to or down to 32 bytes. To do this the function takes a hash
// of the key and cuts it down to 32 bytes.
func hashTo32Bytes(input string) (output string, err error) {

	if len(input) == 0 {
		return "", errors.New("No input supplied")
	}

	hasher := sha256.New()
	hasher.Write([]byte(input))

	stringToSHA256 := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	// Cut the length down to 32 bytes and return.
	return stringToSHA256[:32], nil
}

func main() {

	// Get the amount of arguments from the command line.
	argumentsCount := len(os.Args)

	// Expected usage:
	// encrypt.go -e|-d --key "key here" --value "value here"

	if argumentsCount != 6 {
		fmt.Printf("Usage:\n-e to encrypt, -d to decrypt.\n")
		fmt.Printf("--key \"I am a key\" to load the key.\n")
		fmt.Printf("--value \"I am a text to be encrypted or decrypted\".\n")
		return
	}

	// Set up some flags to check against arguments.
	encrypt := false
	decrypt := false
	key := false
	expectKeyString := 0
	keyString := false
	value := false
	expectValueString := 0
	valueString := false

	// Set the input variables up.
	encryptionFlag := ""
	stringToEncrypt := ""
	encryptionKey := ""

	// Get the arguments from the command line.
	// If any issues are detected, alert the user and exit.
	for index, element := range os.Args {

		if element == "-e" {
			// Ensure that decrypt has not also been set.
			if decrypt == true {
				fmt.Printf("Can't set both -e and -d.\nBye!\n")
				return
			}
			encrypt = true
			encryptionFlag = "-e"

		} else if element == "-d" {
			// Ensure that encrypt has not also been set.
			if encrypt == true {
				fmt.Printf("Can't set both -e and -d.\nBye!\n")
				return
			}
			decrypt = true
			encryptionFlag = "-d"

		} else if element == "--key" {
			key = true
			expectKeyString++

		} else if element == "--value" {
			value = true
			expectValueString++

		} else if expectKeyString == 1 {
			encryptionKey = os.Args[index]
			keyString = true
			expectKeyString = 0

		} else if expectValueString == 1 {
			stringToEncrypt = os.Args[index]
			valueString = true
			expectValueString = 0
		}

		if expectKeyString >= 2 {
			fmt.Printf("Something went wrong, too many keys entered.\bBye!\n")
			return

		} else if expectValueString >= 2 {
			fmt.Printf("Something went wrong, too many keys entered.\bBye!\n")
			return
		}
	}

	// On error, output some useful information.
	if !(encrypt == true || decrypt == true) || key == false || keyString == false || value == false || valueString == false {
		fmt.Printf("Incorrect usage!\n")
		fmt.Printf("---------\n")
		fmt.Printf("-e or -d -> %v\n", (encrypt == true || decrypt == true))
		fmt.Printf("--key -> %v\n", key)
		fmt.Printf("Key string? -> %v\n", keyString)
		fmt.Printf("--value -> %v\n", value)
		fmt.Printf("Value string? -> %v\n", valueString)
		fmt.Printf("---------")
		fmt.Printf("\nUsage:\n-e to encrypt, -d to decrypt.\n")
		fmt.Printf("--key \"I am a key\" to load the key.\n")
		fmt.Printf("--value \"I am a text to be encrypted or decrypted\".\n")
		return
	}

	// Check the encrpytion flag.
	if false == (encryptionFlag == "-e" || encryptionFlag == "-d") {
		fmt.Println("Sorry but the first argument has to be either -e or -d")
		fmt.Println("for either encryption or decryption.")
		return
	}

	if encryptionFlag == "-e" {
		// Encrypt!

		fmt.Printf("Encrypting '%s' with key '%s'\n", stringToEncrypt, encryptionKey)

		encryptedString, _ := encryptString(stringToEncrypt, encryptionKey)

		fmt.Printf("Output: '%s'\n", encryptedString)

	} else if encryptionFlag == "-d" {
		// Decrypt!

		fmt.Printf("Decrypting '%s' with key '%s'\n", stringToEncrypt, encryptionKey)

		decryptedString, _ := decryptString(stringToEncrypt, encryptionKey)

		fmt.Printf("Output: '%s'\n", decryptedString)

	}
}
