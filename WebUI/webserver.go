package webui

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func Webui() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Generate options for the dropdown menu
		options := ""
		for port := 8081; port <= 8090; port++ {
			options += fmt.Sprintf(`<option value="http://localhost:%d/metrics">localhost:%d/metrics</option>`, port, port)
		}

		// HTML form with a dropdown menu
		html := fmt.Sprintf(`
		<html>
		<head>
			<title>Metrics Dashboard</title>
		</head>
		<body>
			<h1>Metrics Dashboard</h1>
			<form action="/fetch" method="get">
				<label for="url">Select URL:</label>
				<select id="url" name="url">
					%s
				</select>
				<button type="submit">Fetch Metrics</button>
			</form>
		</body>
		</html>
		`, options)

		// Send HTML response
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "%s", html)
	})

	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		// Get the selected URL from the form
		url := r.FormValue("url")

		// Fetch data from the selected URL
		resp, err := http.Get(url)
		if err != nil {
			log.Println("Failed to fetch metrics data:", err)
			http.Error(w, "Failed to fetch metrics data", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println("Failed to fetch metrics data. Status code:", resp.StatusCode)
			http.Error(w, "Failed to fetch metrics data", resp.StatusCode)
			return
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read response body:", err)
			http.Error(w, "Failed to read response body", http.StatusInternalServerError)
			return
		}

		// Convert Prometheus format to human-readable format
		humanReadable := prometheusToHumanReadable(string(body))

		// Display the human-readable data as a dashboard
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s", humanReadable)
	})

	// Start the server
	fmt.Println("Server listening on port 8099...")
	log.Fatal(http.ListenAndServe(":8099", nil))
}

// Function to convert Prometheus format to human-readable format
func prometheusToHumanReadable(prometheusData string) string {
	// Split the input by newline to separate metrics
	lines := strings.Split(prometheusData, "\n")

	// Create a buffer to store the human-readable output
	var output strings.Builder

	// Iterate over each metric line
	for _, line := range lines {
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the metric by whitespace to separate labels and values
		parts := strings.Fields(line)

		// Extract metric name and labels
		metricName := parts[0]
		labels := parts[1]

		// Format the metric name and labels
		output.WriteString(fmt.Sprintf("Metric: %s\n", metricName))
		output.WriteString(fmt.Sprintf("Labels: %s\n", labels))

		// Iterate over each value pair
		for _, pair := range parts[2:] {
			// Split the value pair by '=' to separate the value name and value
			valueParts := strings.Split(pair, "=")

			// Format and append the value name and value
			output.WriteString(fmt.Sprintf("%s: %s\n", valueParts[0], valueParts[1]))
		}

		// Add a newline to separate metrics
		output.WriteString("\n")
	}

	// Return the human-readable output
	return output.String()

}
