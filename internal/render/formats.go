package render

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kevinrst/kubot/internal/model"
)

// WriteSARIF renders SARIF 2.1.0 for CI ingestion (e.g. GitHub code
// scanning). Severities map critical→error, warning→warning, note→note;
// suppressed findings carry a suppression entry and never fail anything.
func WriteSARIF(w io.Writer, rep model.Report, version string) error {
	type rule struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		ShortDescription struct {
			Text string `json:"text"`
		} `json:"shortDescription"`
		Help struct {
			Text string `json:"text"`
		} `json:"help,omitempty"`
	}
	type result struct {
		RuleID  string `json:"ruleId"`
		Level   string `json:"level"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
		Locations []struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
			} `json:"physicalLocation"`
		} `json:"locations,omitempty"`
		Suppressions []struct {
			Kind          string `json:"kind"`
			Justification string `json:"justification,omitempty"`
		} `json:"suppressions,omitempty"`
	}
	seen := map[string]bool{}
	var rules []rule
	for _, f := range rep.Issues {
		if seen[f.Reason] {
			continue
		}
		seen[f.Reason] = true
		var r rule
		r.ID = f.Reason
		r.Name = f.Reason
		r.ShortDescription.Text = f.Message
		if f.Recommendation != "" {
			r.Help.Text = f.Recommendation
		}
		rules = append(rules, r)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	var results []result
	for _, f := range rep.Issues {
		var r result
		r.RuleID = f.Reason
		switch f.Severity {
		case model.SeverityCritical:
			r.Level = "error"
		case model.SeverityWarning:
			r.Level = "warning"
		default:
			r.Level = "note"
		}
		r.Message.Text = f.Message
		loc := struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
			} `json:"physicalLocation"`
		}{}
		uri := f.Resource
		if f.Namespace != "" {
			uri = f.Namespace + "/" + uri
		}
		loc.PhysicalLocation.ArtifactLocation.URI = "kubot://" + uri
		r.Locations = []struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
			} `json:"physicalLocation"`
		}{loc}
		if f.Suppressed {
			r.Suppressions = []struct {
				Kind          string `json:"kind"`
				Justification string `json:"justification,omitempty"`
			}{{Kind: "inSource", Justification: f.SuppressionReason}}
		}
		results = append(results, r)
	}
	if results == nil {
		results = []result{}
	}
	doc := map[string]any{
		"version": "2.1.0",
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"runs": []map[string]any{{
			"tool": map[string]any{
				"driver": map[string]any{
					"name":           "kubot",
					"version":        version,
					"informationUri": "https://github.com/kevinrst/kubot",
					"rules":          rules,
				},
			},
			"results": results,
		}},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// WriteJUnit renders JUnit XML for CI test panes (Jenkins, GitLab).
// Criticals and warnings become failures, notes pass silently, suppressed
// findings are skipped — never failed.
func WriteJUnit(w io.Writer, rep model.Report) error {
	type failure struct {
		Message string `xml:"message,attr"`
		Text    string `xml:",chardata"`
	}
	type skipped struct{}
	type kase struct {
		Classname string   `xml:"classname,attr"`
		Name      string   `xml:"name,attr"`
		Failure   *failure `xml:"failure,omitempty"`
		Skipped   *skipped `xml:"skipped,omitempty"`
	}
	failures := 0
	nSkipped := 0
	var cases []kase
	for _, f := range rep.Issues {
		c := kase{Classname: f.Reason, Name: f.Resource}
		switch {
		case f.Suppressed:
			c.Skipped = &skipped{}
			nSkipped++
		case f.Severity == model.SeverityCritical || f.Severity == model.SeverityWarning:
			c.Failure = &failure{Message: f.Message, Text: f.Recommendation}
			failures++
		}
		cases = append(cases, c)
	}
	fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(w, "<testsuites>\n")
	fmt.Fprintf(w, "  <testsuite name=\"kubot\" tests=\"%d\" failures=\"%d\" skipped=\"%d\">\n",
		len(cases), failures, nSkipped)
	for _, c := range cases {
		fmt.Fprintf(w, "    <testcase classname=\"%s\" name=\"%s\"",
			xmlEscape(c.Classname), xmlEscape(c.Name))
		switch {
		case c.Skipped != nil:
			fmt.Fprintf(w, ">\n      <skipped/>\n    </testcase>\n")
		case c.Failure != nil:
			fmt.Fprintf(w, ">\n      <failure message=\"%s\">%s</failure>\n    </testcase>\n",
				xmlEscape(c.Failure.Message), xmlEscape(c.Failure.Text))
		default:
			fmt.Fprintf(w, "/>\n")
		}
	}
	fmt.Fprintf(w, "  </testsuite>\n</testsuites>\n")
	return nil
}

func xmlEscape(s string) string {
	out := ""
	for _, r := range s {
		switch r {
		case '&':
			out += "&amp;"
		case '<':
			out += "&lt;"
		case '>':
			out += "&gt;"
		case '"':
			out += "&quot;"
		default:
			out += string(r)
		}
	}
	return out
}
