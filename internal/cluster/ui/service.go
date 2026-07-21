package ui

import (
	"fmt"
	"io"
	"time"

	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// ClusterDisplayInfo represents cluster information for display purposes
type ClusterDisplayInfo struct {
	Name      string
	Type      string
	Source    string // "local" | "openframe" | "external"
	Context   string // kubeconfig context, when known
	Project   string // GCP project (cloud clusters)
	Status    string
	NodeCount int
	CreatedAt time.Time
	Nodes     []NodeDisplayInfo
}

// NodeDisplayInfo represents node information for display
type NodeDisplayInfo struct {
	Name   string
	Role   string
	Status string
}

// DisplayService handles all cluster-related UI display operations
// This separates presentation concerns from business logic
type DisplayService struct{}

// NewDisplayService creates a new UI display service
func NewDisplayService() *DisplayService {
	return &DisplayService{}
}

// ShowClusterList displays a list of clusters
func (s *DisplayService) ShowClusterList(clusters []ClusterDisplayInfo, out io.Writer) {
	if len(clusters) == 0 {
		fmt.Fprintln(out, "No clusters found.")
		return
	}

	// The SOURCE/CONTEXT/PROJECT columns only appear when the list contains
	// cloud entries — a purely local listing keeps its compact shape.
	hasCloud := false
	for _, c := range clusters {
		if c.Source != "" && c.Source != "local" {
			hasCloud = true
			break
		}
	}

	header := []string{"NAME", "TYPE", "STATUS", "NODES", "CREATED"}
	if hasCloud {
		header = []string{"NAME", "TYPE", "SOURCE", "STATUS", "NODES", "CONTEXT", "PROJECT", "CREATED"}
	}
	tableData := pterm.TableData{header}

	orDash := func(s string) string {
		if s == "" {
			return "—"
		}
		return s
	}
	for _, clusterInfo := range clusters {
		statusColor := sharedUI.GetStatusColor(clusterInfo.Status)
		created := ""
		if !clusterInfo.CreatedAt.IsZero() {
			created = clusterInfo.CreatedAt.Format("2006-01-02 15:04")
		}
		row := []string{
			pterm.Bold.Sprint(clusterInfo.Name),
			clusterInfo.Type,
			statusColor(clusterInfo.Status),
			fmt.Sprintf("%d", clusterInfo.NodeCount),
			created,
		}
		if hasCloud {
			row = []string{
				pterm.Bold.Sprint(clusterInfo.Name),
				clusterInfo.Type,
				orDash(clusterInfo.Source),
				statusColor(clusterInfo.Status),
				fmt.Sprintf("%d", clusterInfo.NodeCount),
				orDash(clusterInfo.Context),
				orDash(clusterInfo.Project),
				created,
			}
		}
		tableData = append(tableData, row)
	}

	// Use pterm table for better formatting - but write to the provided writer
	table := pterm.DefaultTable.WithHasHeader().WithData(tableData).WithWriter(out)
	if err := table.Render(); err != nil {
		// Fallback to simple formatting if pterm fails
		for i, row := range tableData {
			if i == 0 {
				// Header row
				fmt.Fprintf(out, "%-17s %-8s %-10s %-6s %s\n", row[0], row[1], row[2], row[3], row[4])
				continue
			}
			// Data rows - need to account for styled text by using different spacing
			fmt.Fprintf(out, "%-17s %-8s %-10s %-6s %s\n",
				pterm.RemoveColorFromString(row[0]), // Remove color codes for alignment
				row[1],
				pterm.RemoveColorFromString(row[2]), // Remove color codes for alignment
				row[3],
				row[4])
		}
	}
}
