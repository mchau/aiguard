package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/audit"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/requirements"
)

var reqCmd = &cobra.Command{
	Use:   "req",
	Short: "Manage requirements for this work item",
}

var reqIngestFile string
var reqIngestPaste bool
var reqIngestJira string

var reqIngestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest a requirement (ticket, file, or paste)",
	RunE:  runReqIngest,
}

var reqStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show ingested requirement status",
	RunE:  runReqStatus,
}

func init() {
	reqIngestCmd.Flags().StringVar(&reqIngestFile, "file", "", "path to requirement file")
	reqIngestCmd.Flags().BoolVar(&reqIngestPaste, "paste", false, "read requirement from stdin")
	reqIngestCmd.Flags().StringVar(&reqIngestJira, "jira", "", "Jira issue ID (not yet implemented)")
	reqCmd.AddCommand(reqIngestCmd)
	reqCmd.AddCommand(reqStatusCmd)
	rootCmd.AddCommand(reqCmd)
}

func runReqIngest(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	rawMD := project.RawMD(root)

	switch {
	case reqIngestFile != "":
		if err := requirements.IngestFile(reqIngestFile, rawMD); err != nil {
			return err
		}
		fmt.Printf("Ingested requirement from %s → %s\n", reqIngestFile, rawMD)

	case reqIngestPaste:
		fmt.Fprintln(os.Stderr, "Paste requirement text below, then press Ctrl+D:")
		if err := requirements.IngestStdin(rawMD, os.Stdin); err != nil {
			return err
		}
		fmt.Printf("Ingested requirement from stdin → %s\n", rawMD)

	case reqIngestJira != "":
		src := requirements.JiraSource{}
		_, err := src.Fetch(reqIngestJira)
		return fmt.Errorf("jira ingest: %w — use --file or --paste instead", err)

	default:
		return fmt.Errorf("specify --file, --paste, or --jira")
	}

	_ = audit.Append(root, audit.Event{Command: "req ingest", Inputs: []string{reqIngestFile}, Outputs: []string{rawMD}, Status: "success"})
	fmt.Println("Next step: run 'aiguard digest create' to extract acceptance criteria.")
	return nil
}

func runReqStatus(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	rawMD := project.RawMD(root)
	info, err := os.Stat(rawMD)
	if os.IsNotExist(err) {
		fmt.Println("No requirement ingested. Run 'aiguard req ingest' first.")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("raw.md: %s (%d bytes)\n", rawMD, info.Size())
	return nil
}
