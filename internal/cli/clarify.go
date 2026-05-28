package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/mchau/aiguard/internal/project"
	"github.com/mchau/aiguard/internal/requirements"
)

var clarifyCmd = &cobra.Command{
	Use:   "clarify",
	Short: "Manage clarification questions for the current requirement",
	RunE:  runClarify,
}

var clarifyAnswerID string
var clarifyAnswerText string
var clarifyAnswerAssume bool

var clarifyAnswerCmd = &cobra.Command{
	Use:   "answer",
	Short: "Answer or assume a clarification question",
	RunE:  runClarifyAnswer,
}

func init() {
	clarifyAnswerCmd.Flags().StringVar(&clarifyAnswerID, "id", "", "question ID (e.g. Q1)")
	clarifyAnswerCmd.Flags().StringVar(&clarifyAnswerText, "text", "", "answer text")
	clarifyAnswerCmd.Flags().BoolVar(&clarifyAnswerAssume, "assume", false, "mark as assumed rather than answered")
	_ = clarifyAnswerCmd.MarkFlagRequired("id")
	_ = clarifyAnswerCmd.MarkFlagRequired("text")
	clarifyCmd.AddCommand(clarifyAnswerCmd)
	rootCmd.AddCommand(clarifyCmd)
}

func runClarify(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	questionsPath := project.OpenQuestionsMD(root)
	data, err := os.ReadFile(questionsPath)
	if os.IsNotExist(err) {
		fmt.Println("No open questions found. Run 'aiguard digest create' first.")
		return nil
	}
	if err != nil {
		return err
	}

	// Load questions from JSON sidecar if present
	questions := loadQuestionsJSON(project.OpenQuestionsMD(root) + ".json")
	if len(questions) == 0 {
		fmt.Println("Questions file:")
		fmt.Println(string(data))
		return nil
	}

	blocking := 0
	for _, q := range questions {
		if q.Blocking && q.Status == "open" {
			blocking++
		}
	}

	if blocking > 0 {
		fmt.Printf("WARN: %d blocking question(s) are open. Downstream commands will be blocked until answered.\n", blocking)
	}

	for _, q := range questions {
		status := q.Status
		if status == "" {
			status = "open"
		}
		blocking := ""
		if q.Blocking {
			blocking = " [BLOCKING]"
		}
		fmt.Printf("  %s%s: %s (status: %s)\n", q.ID, blocking, q.Text, status)
	}
	return nil
}

func runClarifyAnswer(cmd *cobra.Command, args []string) error {
	root, err := requireRoot()
	if err != nil {
		return err
	}

	qPath := project.OpenQuestionsMD(root) + ".json"
	questions := loadQuestionsJSON(qPath)

	found := false
	for i := range questions {
		if questions[i].ID == clarifyAnswerID {
			questions[i].Answer = clarifyAnswerText
			if clarifyAnswerAssume {
				questions[i].Status = "assumed"
			} else {
				questions[i].Status = "answered"
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("question %q not found; run 'aiguard clarify' to list questions", clarifyAnswerID)
	}

	data, err := json.MarshalIndent(questions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(qPath, data, 0o644); err != nil {
		return err
	}

	// Rewrite questions MD
	if err := os.WriteFile(project.OpenQuestionsMD(root),
		[]byte(requirements.RenderQuestions(questions)), 0o644); err != nil {
		return err
	}

	action := "answered"
	if clarifyAnswerAssume {
		action = "assumed"
	}
	fmt.Printf("Question %s marked as %s.\n", clarifyAnswerID, action)
	return nil
}

func loadQuestionsJSON(path string) []requirements.ClarificationQuestion {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var questions []requirements.ClarificationQuestion
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil
	}
	return questions
}
