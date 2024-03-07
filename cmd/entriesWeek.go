package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

const (
	blue    = lipgloss.Color("#8EB8E5")
	emerald = lipgloss.Color("#23CE6B")
	gray    = lipgloss.Color("#61707D")
)

var weekCmd = &cobra.Command{
	Use:   "week",
	Short: "Get entries for this week",
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := os.Getenv("TIMECAMP_API_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Missing TIMECAMP_API_TOKEN environment variable")
			return
		}

		current := time.Now()
		year, month, day := current.Date()

		url := fmt.Sprintf("https://app.timecamp.com/third_party/api/logged_time_in_week?day=%d-%02d-%02d", year, month, day)

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+apiToken)

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
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

		var entries map[string]int64
		err := json.Unmarshal(body, &entries)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		type Response struct {
			Date    string
			Seconds int64
		}

		var entryList []Response
		for date, seconds := range entries {
			entryList = append(entryList, Response{Date: date, Seconds: seconds})
		}
		sort.Slice(entryList, func(i, j int) bool {
			return entryList[i].Date < entryList[j].Date
		})

		var rowsDate []string
		var rowsDuration []string
		for _, entry := range entryList {
			date, err := time.Parse("2006-01-02", entry.Date)
			if err != nil {
				fmt.Printf("Error parsing date: %s\n", err)
				continue
			}
			duration := time.Duration(entry.Seconds) * time.Second
			rowsDate = append(rowsDate, date.Format("Mon, 02 Jan"))
			rowsDuration = append(rowsDuration, duration.String())
		}
		re := lipgloss.NewRenderer(os.Stdout)
		var (
			HeaderStyle = re.NewStyle().Padding(0, 1).Foreground(blue).Bold(true).Align(lipgloss.Center)
			CellStyle   = re.NewStyle().Padding(0, 1)
		)
		rows := [][]string{
			rowsDuration,
		}
		t := table.New().
			Border(lipgloss.NormalBorder()).
			Headers(rowsDate...).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == 0 {
					if rowsDate[col] == current.Format("Mon, 02 Jan") {
						return HeaderStyle.Copy().Foreground(emerald)
					}
					return HeaderStyle
				}
				if rows[row-1][col] == "0s" {
					return CellStyle.Copy().Foreground(gray)
				}
				return CellStyle
			})
		fmt.Println(t)
	},
}

func init() {
	entriesCmd.AddCommand(weekCmd)
}
