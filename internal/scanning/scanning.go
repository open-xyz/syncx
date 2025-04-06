package scanning

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// ScanResult represents the output of a security scan
type ScanResult struct {
	ProjectID   int       `json:"project_id"`
	ScanDate    time.Time `json:"scan_date"`
	VulnCount   int       `json:"vuln_count"`
	RawOutput   string    `json:"raw_output"`
	Severity    string    `json:"severity"` // e.g., "HIGH", "MEDIUM", "LOW"
	Recommended string    `json:"recommended,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// IsTrivyInstalled checks if Trivy is installed on the system
func IsTrivyInstalled() bool {
	_, err := exec.LookPath("trivy")
	return err == nil
}

// GetTrivyInstallInstructions returns installation instructions for Trivy
func GetTrivyInstallInstructions() string {
	return `Trivy is not installed. Please install it using:
    
    # For Arch Linux/Manjaro
    sudo pacman -S trivy
    
    # For Ubuntu/Debian
    sudo apt-get install trivy
    
    # For macOS
    brew install trivy
    
    For more information, visit: https://aquasecurity.github.io/trivy/latest/getting-started/installation/`
}

// ScanProject scans a project directory for vulnerabilities using Trivy
func ScanProject(projectID int, path string) (ScanResult, error) {
	result := ScanResult{
		ProjectID: projectID,
		ScanDate:  time.Now(),
	}

	// Check if Trivy is installed
	if !IsTrivyInstalled() {
		errMsg := "Trivy is not installed. Cannot perform security scan."
		log.Printf("Scan failed for project %d: %s", projectID, errMsg)
		result.Error = errMsg
		result.Recommended = GetTrivyInstallInstructions()
		return result, fmt.Errorf(errMsg)
	}

	// Run Trivy in filesystem scanning mode with JSON output
	cmd := exec.Command("trivy", "fs", "--format", "json", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Trivy scan error: %v", err)
		// Still store the output even if there's an error
		result.RawOutput = string(output)
		result.Error = fmt.Sprintf("Scan error: %v", err)
		return result, err
	}

	result.RawOutput = string(output)
	
	// Parse JSON output to get vulnerability count
	var trivyResult map[string]interface{}
	if err := json.Unmarshal(output, &trivyResult); err != nil {
		log.Printf("Failed to parse Trivy output: %v", err)
		result.Error = fmt.Sprintf("Failed to parse scan results: %v", err)
	} else {
		// Extract vulnerability count and severity from Trivy results
		if results, ok := trivyResult["Results"].([]interface{}); ok {
			var totalVulns int
			var highestSeverity string
			
			for _, res := range results {
				if resultMap, ok := res.(map[string]interface{}); ok {
					if vulns, ok := resultMap["Vulnerabilities"].([]interface{}); ok {
						totalVulns += len(vulns)
						
						// Determine highest severity
						for _, vuln := range vulns {
							if vulnMap, ok := vuln.(map[string]interface{}); ok {
								if sev, ok := vulnMap["Severity"].(string); ok {
									if isSeverityHigher(sev, highestSeverity) {
										highestSeverity = sev
									}
								}
							}
						}
					}
				}
			}
			
			result.VulnCount = totalVulns
			result.Severity = highestSeverity
			
			// Add basic recommendations based on severity
			if totalVulns > 0 {
				result.Recommended = getRecommendation(highestSeverity)
			} else {
				result.Recommended = "No vulnerabilities found."
			}
		}
	}

	log.Printf("Scan completed for project %d: found %d vulnerabilities with highest severity %s",
		projectID, result.VulnCount, result.Severity)
	
	return result, nil
}

// isSeverityHigher checks if severity a is higher than severity b
func isSeverityHigher(a, b string) bool {
	severityRank := map[string]int{
		"CRITICAL": 4,
		"HIGH":     3,
		"MEDIUM":   2,
		"LOW":      1,
		"UNKNOWN":  0,
		"":         -1,
	}
	
	rankA := severityRank[a]
	rankB := severityRank[b]
	
	return rankA > rankB
}

// getRecommendation returns a basic recommendation based on severity
func getRecommendation(severity string) string {
	switch severity {
	case "CRITICAL":
		return "Immediate action required. Update vulnerable dependencies as soon as possible."
	case "HIGH":
		return "High priority fix needed. Update affected components soon."
	case "MEDIUM":
		return "Update affected components during next maintenance cycle."
	case "LOW":
		return "Low risk. Consider updating during regular maintenance."
	default:
		return "Unknown severity. Review scan details for more information."
	}
} 