package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var entriesFromFlag string

var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "Get today's entries",
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := os.Getenv("TIMECAMP_API_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Missing TIMECAMP_API_TOKEN environment variable")
			return
		}

		fromDate, err := time.Parse("2006-01-02", entriesFromFlag)
		if err != nil {
			fmt.Println("Error parsing date:", err)
			return
		}
		year, month, day := fromDate.Date()

		url := fmt.Sprintf("https://app.timecamp.com/third_party/api/entries?from=%d-%02d-%02d&to=%d-%02d-%02d", year, month, day, year, month, day)

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+apiToken)

		res, _ := http.DefaultClient.Do(req)

		body, _ := io.ReadAll(res.Body)
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			var errorMessage ErrorMessage
			err := json.Unmarshal([]byte(body), &errorMessage)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Println("Error:", errorMessage.Message)
			return
		}

		type Response struct {
			ID               int64  `json:"id"`
			Duration         string `json:"duration"`
			UserID           string `json:"user_id"`
			UserName         string `json:"user_name"`
			TaskID           string `json:"task_id"`
			TaskNote         string `json:"task_note"`
			LastModify       string `json:"last_modify"`
			Date             string `json:"date"`
			StartTime        string `json:"start_time"`
			EndTime          string `json:"end_time"`
			Locked           string `json:"locked"`
			Name             string `json:"name"`
			AddonsExternalID string `json:"addons_external_id"`
			Billable         int    `json:"billable"`
			InvoiceID        string `json:"invoiceId"`
			Color            string `json:"color"`
			Description      string `json:"description"`
		}

		var entries []Response
		err = json.Unmarshal(body, &entries)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		loc, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		styledDate := lipgloss.NewStyle().Foreground(blue).Bold(true)
		fmt.Println(" " + styledDate.Render(fromDate.Format("Monday, January 02, 2006")))

		re := lipgloss.NewRenderer(os.Stdout)
		var (
			HeaderStyle = re.NewStyle().Padding(0, 1).Foreground(blue).Bold(true).Align(lipgloss.Center)
			CellStyle   = re.NewStyle().Padding(0, 1)
		)
		t := table.New().
			Border(lipgloss.NormalBorder()).
			Headers("Entry ID", "Task", "From", "To", "Duration").
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					return HeaderStyle
				}
				switch col {
				case 1:
					color := lipgloss.Color(entries[row-1].Color)
					return CellStyle.Copy().Foreground(color)
				case 4:
					return CellStyle.Copy().Bold(true).Align(lipgloss.Right)
				default:
					return CellStyle
				}
			})
		var totalDuration time.Duration
		for _, entry := range entries {
			seconds, _ := strconv.ParseInt(entry.Duration, 10, 64)
			duration := "0s"
			if seconds == 0 && entry.StartTime == entry.EndTime {
				givenTime, err := time.ParseInLocation("2006-01-02 15:04:05", entry.Date+" "+entry.StartTime, loc)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				elapsedTime := time.Since(givenTime)
				totalDuration += elapsedTime.Round(time.Second)
				duration = elapsedTime.Round(time.Second).String()
			} else {
				elapsedTime := time.Duration(seconds) * time.Second
				totalDuration += elapsedTime
				duration = elapsedTime.String()
			}
			t.Row(fmt.Sprintf("%d", entry.ID), entry.Name, entry.StartTime, entry.EndTime, duration)
		}
		fmt.Println(t)
		styledTotal := lipgloss.NewStyle().Bold(true).Align(lipgloss.Right)
		tableWidth := lipgloss.Width(t.Render()) - 2
		totalBlock := lipgloss.PlaceHorizontal(tableWidth, lipgloss.Right, styledTotal.Render(fmt.Sprintf("Total %s", totalDuration)))
		fmt.Println(totalBlock)
	},
}

func init() {
	current := time.Now().Local()
	year, month, day := current.Date()
	rootCmd.AddCommand(entriesCmd)
	entriesCmd.Flags().StringVarP(&entriesFromFlag, "from", "", fmt.Sprintf("%d-%02d-%02d", year, month, day), fmt.Sprintf("From date format: %d-%02d-%02d", year, month, day))
}
