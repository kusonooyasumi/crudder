package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

func makeRequest(method, baseUrl, endpoint string, outputFile *os.File, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement counter when the goroutine completes

	// Combine baseUrl and endpoint to form the complete URL
	url := baseUrl + endpoint

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		handleError(fmt.Sprintf("Failed to create request for %s: %v", url, err), outputFile)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		handleError(fmt.Sprintf("Failed to make %s request to %s: %v", method, url, err), outputFile)
		return
	}
	defer resp.Body.Close()

	// Prepare the result string
	result := fmt.Sprintf("%s request to %s: %d", method, url, resp.StatusCode)
	result = removePattern(result)

	// Print the result
	fmt.Println(result)
	if outputFile != nil {
		writeToFile(result, outputFile)
	}
}

func handleError(message string, outputFile *os.File) {
	message = removePattern(message)
	fmt.Println(message)
	if outputFile != nil {
		writeToFile(message, outputFile)
	}
}

func removePattern(input string) string {
	re := regexp.MustCompile(`request to .*:`)
	return re.ReplaceAllString(input, "")
}

func writeToFile(data string, outputFile *os.File) {
	_, err := outputFile.WriteString(data + "\n")
	if err != nil {
		fmt.Println("Failed to write to file:", err)
	}
}

func readFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func parseSubdomains(subdomainsInput string) []string {
	// If subdomains are given as a comma-separated list, split them
	return strings.Split(subdomainsInput, ",")
}

func getSubdomainsFromFile(filePath string) ([]string, error) {
	return readFile(filePath)
}

func main() {
	// Command line arguments
	var methods string
	var subdomainsInput string
	var subdomainsFile string
	var endpointsFile string
	var outputFile string
	var maxConcurrent int

	// Default concurrency
	defaultConcurrency := 50

	// Prompt for user input
	fmt.Println("Enter HTTP methods (c=POST, r=GET, u=PUT, d=DELETE): ")
	fmt.Scanln(&methods)
	fmt.Println("Enter subdomain(s) (comma-separated for multiple subdomains, e.g. 'subdomain1,subdomain2') or leave blank: ")
	fmt.Scanln(&subdomainsInput)
	fmt.Println("Enter the file path containing subdomains (one per line) if using -sf: ")
	fmt.Scanln(&subdomainsFile)
	fmt.Println("Enter the file path containing API endpoints (one per line): ")
	fmt.Scanln(&endpointsFile)
	fmt.Println("Enter the output file path (optional): ")
	fmt.Scanln(&outputFile)
	fmt.Println("Enter number of concurrent requests (default 50): ")
	fmt.Scanln(&maxConcurrent)

	// If no value is provided for concurrency, use default
	if maxConcurrent == 0 {
		maxConcurrent = defaultConcurrency
	}

	// Determine which subdomains to use (priority: -sf > -s)
	var subdomains []string
	var err error
	if subdomainsFile != "" {
		// Read subdomains from file if -sf is provided
		subdomains, err = getSubdomainsFromFile(subdomainsFile)
		if err != nil {
			fmt.Println("Error reading subdomains from file:", err)
			return
		}
	} else if subdomainsInput != "" {
		// Otherwise, use the comma-separated list from -s
		subdomains = parseSubdomains(subdomainsInput)
	}

	// Map the method options
	methodMap := map[rune]string{
		'u': "PUT",
		'r': "GET",
		'd': "DELETE",
		'c': "POST",
	}

	var selectedMethods []string
	for _, method := range methods {
		if val, exists := methodMap[method]; exists {
			selectedMethods = append(selectedMethods, val)
		}
	}

	if len(selectedMethods) == 0 {
		fmt.Println("No valid methods selected. Please use 'c', 'r', 'u', or 'd'.")
		return
	}

	// Read the list of endpoints from the file
	endpoints, err := readFile(endpointsFile)
	if err != nil {
		fmt.Println("Error reading endpoints from file:", err)
		return
	}

	// Open the output file if it's provided
	var outputFilePtr *os.File
	if outputFile != "" {
		outputFilePtr, err = os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer outputFilePtr.Close()
	}

	// Create a WaitGroup to synchronize the concurrent requests
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent) // Limit concurrent requests

	// Make requests for each combination of subdomain and endpoint
	for _, subdomain := range subdomains {
		for _, endpoint := range endpoints {
			// Manage concurrency with a semaphore
			semaphore <- struct{}{} // Acquire a slot
			wg.Add(1)

			go func(subdomain, endpoint string) {
				defer func() { <-semaphore }() // Release the slot when done
				fmt.Printf("Subdomain: %s, Endpoint: %s\n", subdomain, endpoint)
				for _, method := range selectedMethods {
					makeRequest(method, subdomain, endpoint, outputFilePtr, &wg)
				}
			}(subdomain, endpoint)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
