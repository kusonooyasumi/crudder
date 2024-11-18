package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

const banner = `
  /$$$$$$  /$$$$$$$  /$$   /$$ /$$$$$$$        /$$                    
 /$$__  $$| $$__  $$| $$  | $$| $$__  $$      | $$                    
| $$  \__/| $$  \ $$| $$  | $$| $$  \ $$  /$$$$$$$  /$$$$$$   /$$$$$$ 
| $$      | $$$$$$$/| $$  | $$| $$  | $$ /$$__  $$ /$$__  $$ /$$__  $$
| $$      | $$__  $$| $$  | $$| $$  | $$| $$  | $$| $$$$$$$$| $$  \__/
| $$    $$| $$  \ $$| $$  | $$| $$  | $$| $$  | $$| $$_____/| $$      
|  $$$$$$/| $$  | $$|  $$$$$$/| $$$$$$$/|  $$$$$$$|  $$$$$$$| $$      
 \______/ |__/  |__/ \______/ |_______/  \_______/ \_______/|__/       by oyasumi(@kusonooyasumi)
`

// Result struct to store request results
type Result struct {
	endpoint    string
	subdomain   string
	method      string
	statusCode  int
	error       string
}

func ensureProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

func makeRequest(method, baseUrl, endpoint string, resultChan chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	// Ensure baseUrl has protocol and doesn't end with slash
	baseUrl = ensureProtocol(strings.TrimSuffix(baseUrl, "/"))
	// Ensure endpoint starts with slash
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	url := baseUrl + endpoint
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		resultChan <- Result{
			endpoint:  endpoint,
			subdomain: baseUrl,
			method:    method,
			error:     fmt.Sprintf("Failed to create request: %v", err),
		}
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		resultChan <- Result{
			endpoint:  endpoint,
			subdomain: baseUrl,
			method:    method,
			error:     fmt.Sprintf("Request failed: %v", err),
		}
		return
	}
	defer resp.Body.Close()

	resultChan <- Result{
		endpoint:   endpoint,
		subdomain:  baseUrl,
		method:     method,
		statusCode: resp.StatusCode,
	}
}

func writeResults(results []Result, outputFile *os.File) {
	// Group results by endpoint
	endpointMap := make(map[string][]Result)
	for _, result := range results {
		endpointMap[result.endpoint] = append(endpointMap[result.endpoint], result)
	}

	// Write results to both console and file
	for endpoint, endpointResults := range endpointMap {
		output := fmt.Sprintf("\nEndpoint: %s\n", endpoint)
		output += "----------------------------------------\n"
		
		// Group by subdomain
		subdomainMap := make(map[string][]Result)
		for _, result := range endpointResults {
			subdomainMap[result.subdomain] = append(subdomainMap[result.subdomain], result)
		}

		for subdomain, subResults := range subdomainMap {
			output += fmt.Sprintf("  Subdomain: %s\n", subdomain)
			for _, result := range subResults {
				if result.error != "" {
					output += fmt.Sprintf("    %s: Error - %s\n", result.method, result.error)
				} else {
					output += fmt.Sprintf("    %s: %d\n", result.method, result.statusCode)
				}
			}
		}
		output += "----------------------------------------\n"

		fmt.Print(output)
		if outputFile != nil {
			outputFile.WriteString(output)
		}
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
		line := strings.TrimSpace(scanner.Text())
		if line != "" { // Skip empty lines
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func parseSubdomains(subdomainsInput string) []string {
	subdomains := strings.Split(subdomainsInput, ",")
	var result []string
	for _, s := range subdomains {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func getSubdomainsFromFile(filePath string) ([]string, error) {
	return readFile(filePath)
}

func main() {
	fmt.Println(banner)

	// Define command-line flags
	methods := flag.String("m", "", "HTTP methods (c=POST, r=GET, u=PUT, d=DELETE)")
	subdomainsInput := flag.String("s", "", "Comma-separated list of subdomains")
	subdomainsFile := flag.String("sf", "", "File containing subdomains (one per line)")
	endpointsFile := flag.String("e", "", "File containing API endpoints (one per line)")
	outputFile := flag.String("o", "", "Output file path (optional)")
	maxConcurrent := flag.Int("r", 50, "Number of concurrent requests (default 50)")

	// Parse command-line flags
	flag.Parse()

	// Validate required flags
	if *methods == "" {
		fmt.Println("Error: HTTP methods (-m) is required")
		flag.Usage()
		os.Exit(1)
	}

	if *endpointsFile == "" {
		fmt.Println("Error: Endpoints file (-e) is required")
		flag.Usage()
		os.Exit(1)
	}

	if *subdomainsInput == "" && *subdomainsFile == "" {
		fmt.Println("Error: Either subdomains (-s) or subdomains file (-sf) is required")
		flag.Usage()
		os.Exit(1)
	}

	// Determine which subdomains to use (priority: -sf > -s)
	var subdomains []string
	var err error
	if *subdomainsFile != "" {
		subdomains, err = getSubdomainsFromFile(*subdomainsFile)
		if err != nil {
			fmt.Println("Error reading subdomains from file:", err)
			os.Exit(1)
		}
	} else if *subdomainsInput != "" {
		subdomains = parseSubdomains(*subdomainsInput)
	}

	if len(subdomains) == 0 {
		fmt.Println("Error: No valid subdomains provided")
		os.Exit(1)
	}

	// Map the method options
	methodMap := map[rune]string{
		'u': "PUT",
		'r': "GET",
		'd': "DELETE",
		'c': "POST",
	}

	var selectedMethods []string
	for _, method := range *methods {
		if val, exists := methodMap[method]; exists {
			selectedMethods = append(selectedMethods, val)
		}
	}

	if len(selectedMethods) == 0 {
		fmt.Println("No valid methods selected. Please use 'c', 'r', 'u', or 'd'.")
		os.Exit(1)
	}

	// Read the list of endpoints from the file
	endpoints, err := readFile(*endpointsFile)
	if err != nil {
		fmt.Println("Error reading endpoints from file:", err)
		os.Exit(1)
	}

	if len(endpoints) == 0 {
		fmt.Println("Error: No valid endpoints found in file")
		os.Exit(1)
	}

	// Open the output file if it's provided
	var outputFilePtr *os.File
	if *outputFile != "" {
		outputFilePtr, err = os.Create(*outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}
		defer outputFilePtr.Close()
	}

	// Create a WaitGroup and result channel to collect results
	var wg sync.WaitGroup
	resultChan := make(chan Result, len(subdomains)*len(endpoints)*len(selectedMethods))
	semaphore := make(chan struct{}, *maxConcurrent)

	// Calculate total number of requests
	totalRequests := len(subdomains) * len(endpoints) * len(selectedMethods)
	wg.Add(totalRequests)

	// Make requests for each combination of subdomain and endpoint
	for _, subdomain := range subdomains {
		for _, endpoint := range endpoints {
			semaphore <- struct{}{}

			go func(subdomain, endpoint string) {
				defer func() { <-semaphore }()
				fmt.Printf("Testing: %s%s\n", subdomain, endpoint)
				for _, method := range selectedMethods {
					makeRequest(method, subdomain, endpoint, resultChan, &wg)
				}
			}(subdomain, endpoint)
		}
	}

	// Start a goroutine to close the result channel when all requests are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []Result
	for result := range resultChan {
		results = append(results, result)
	}

	// Write organized results to output
	writeResults(results, outputFilePtr)
}
