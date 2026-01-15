package security

// Severity levels for security issues
type Severity int

const (
	SeverityLow    Severity = 1
	SeverityMedium Severity = 2
	SeverityHigh   Severity = 3
)

// SecurityIssue represents a single security issue with severity
type SecurityIssue struct {
	Severity Severity
	Pattern  string
	File     string
	Line     int
	Content  string
}

// SeverityCheckResult holds results with severity information
type SeverityCheckResult struct {
	Passed bool
	Issues []SecurityIssue
}

// PatternSeverity maps pattern names to severity levels
var PatternSeverity = map[string]Severity{
	"AWS Access Key ID":    SeverityHigh,
	"AWS Secret Key":       SeverityHigh,
	"Private Key":          SeverityHigh,
	"Password Assignment":  SeverityMedium,
	"API Key Assignment":   SeverityMedium,
	"Secret Assignment":    SeverityMedium,
	"MongoDB Connection":   SeverityHigh,
	"Postgres Connection":  SeverityHigh,
}

// GetPatternSeverity returns the severity for a pattern name
func GetPatternSeverity(patternName string) Severity {
	if sev, ok := PatternSeverity[patternName]; ok {
		return sev
	}
	return SeverityLow // Default to low if unknown
}

// SeverityString returns string representation of severity
func SeverityString(s Severity) string {
	switch s {
	case SeverityHigh:
		return "HIGH"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityLow:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

// CheckProjectWithSeverity performs security check with severity classification
func CheckProjectWithSeverity(dir string, mode string) SeverityCheckResult {
	result := SeverityCheckResult{Passed: true}

	// Scan for secrets
	scanResults, _ := ScanDirectory(dir)

	for _, file := range scanResults {
		for _, match := range file.Matches {
			severity := GetPatternSeverity(match.Pattern)

			// Adjust severity based on mode
			switch mode {
			case "strict":
				// In strict mode, all issues are high severity
				severity = SeverityHigh
			case "relaxed":
				// In relaxed mode, downgrade all to low (warnings only)
				severity = SeverityLow
			}
			// "default" mode uses the natural severity

			issue := SecurityIssue{
				Severity: severity,
				Pattern:  match.Pattern,
				File:     file.Path,
				Line:     match.Line,
				Content:  match.Content,
			}

			result.Issues = append(result.Issues, issue)

			// Only high severity issues cause failure in default mode
			if mode == "default" {
				if severity == SeverityHigh {
					result.Passed = false
				}
			} else if mode == "strict" {
				// Any issue fails in strict mode
				result.Passed = false
			}
			// In relaxed mode, nothing fails (just warnings)
		}
	}

	return result
}
