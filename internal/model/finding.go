package model

const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityNote     = "note"
)

var SeverityOrder = map[string]int{
	SeverityCritical: 0,
	SeverityWarning:  1,
	SeverityNote:     2,
}

func RankSeverity(s string) int {
	if r, ok := SeverityOrder[s]; ok {
		return r
	}
	return len(SeverityOrder)
}

type Finding struct {
	Severity       string         `json:"severity"`
	Resource       string         `json:"resource"`
	Namespace      string         `json:"namespace,omitempty"`
	Reason         string         `json:"reason"` // snake_case, stable id
	Message        string         `json:"message"`
	Evidence       map[string]any `json:"evidence,omitempty"`
	Recommendation string         `json:"recommendation,omitempty"`
}

type ClusterInfo struct {
	Context   string `json:"context,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Report struct {
	SchemaVersion string      `json:"schema_version"`
	Cluster       ClusterInfo `json:"cluster,omitempty"`
	Status        string      `json:"status"` // ok|warning|critical
	Issues        []Finding   `json:"issues"`
	Checked       []string    `json:"checked,omitempty"`
	Degraded      []string    `json:"degraded,omitempty"` // collectors that failed; partial data
}

func OverallStatus(findings []Finding) string {
	status := "ok"
	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			return "critical"
		case SeverityWarning:
			status = "warning"
		}
	}
	return status
}
