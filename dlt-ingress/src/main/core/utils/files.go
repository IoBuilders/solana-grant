package utils

import (
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// Separator defines a universally supported string to separate the timestamp and fileName.
var Separator = "-"

// AddTimeStampToFileName prepends a UNIX timestamp (current time in seconds since epoch) to the given file name.
// This ensures unique identifiers for every file.
func AddTimeStampToFileName(fileName string) string {
	return strconv.FormatInt(time.Now().Unix(), 10) + Separator + fileName
}

// RemoveTimestampFromFilename extracts and removes the timestamp prefix from a file name
// and returns the original file name.
//
// Parameters:
//   - fileName (string): The file name with the prepended timestamp.
//
// Returns:
//   - string: The original file name without the timestamp.
//
// Notes:
//   - Assumes the fileName contains a valid timestamp prefix followed by the Separator.
func RemoveTimestampFromFilename(fileName string) string {
	// Split the string by the Separator and return the second part (the original file name).
	parts := strings.Split(fileName, Separator)
	if len(parts) > 1 {
		return strings.Join(parts[1:], Separator) // In case the fileName contains extra separators
	}
	return fileName // Return unchanged if no separator was found
}

// SHA3HashFile computes the Keccak-256 hash (legacy SHA3-256 variant) of the provided file data.
// It then formats the resulting hash as a string and removes the "0x" prefix.
//
// Parameters:
//   - fileData: A byte slice containing the data of the file to hash.
//
// Returns:
//   - A string representation of the computed Keccak-256 hash of the file data.
func SHA3HashFile(fileData []byte) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(fileData)
	hashString := common.BytesToHash(hash.Sum(nil)).String()
	return strings.TrimPrefix(hashString, "0x")
}

// CleanFileName sanitizes a fileName by allowing only alphanumeric characters,
// underscores, hyphens, and dots, and replaces disallowed characters with an underscore.
//
// Parameters:
//
//	fileName - The string representing the original fileName.
//
// Returns:
//
//	A sanitized fileName containing only safe characters.
func CleanFileName(fileName string) (string, error) {
	fileName = strings.ReplaceAll(fileName, " ", "_")

	re, err := regexp.Compile(`[^a-zA-Z0-9_.-]`)
	if err != nil {
		return fileName, err
	}

	return re.ReplaceAllString(fileName, ""), nil
}

// ValidateURL checks if the provided URL string has a valid format.
// Returns true if the URL is valid, otherwise false.
func ValidateURL(inputURL string) bool {
	parsedURL, err := url.ParseRequestURI(inputURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

// ExtractFileNameFromURL extracts the fileName from a given URL string.
// Returns the fileName and an error if the URL is invalid.
func ExtractFileNameFromURL(fileUrl string) (string, error) {
	parsedUrl, err := url.Parse(fileUrl)
	if err != nil {
		return "", err
	}

	fileName := path.Base(parsedUrl.Path)
	return fileName, nil
}

// CalculateFileSizeFromBytes calculates the file size from a byte slice.
func CalculateFileSizeFromBytes(fileData []byte) int64 {
	return int64(len(fileData))
}
