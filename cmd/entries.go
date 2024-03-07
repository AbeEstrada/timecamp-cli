package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

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

		fmt.Println(fromDate.Format("Mon, 02 Jan"))

		var totalDuration time.Duration
		for _, entry := range entries {
			seconds, _ := strconv.ParseInt(entry.Duration, 10, 64)
			if seconds == 0 && entry.StartTime == entry.EndTime {
				givenTime, err := time.ParseInLocation("2006-01-02 15:04:05", entry.Date+" "+entry.StartTime, loc)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				elapsedTime := time.Since(givenTime)
				totalDuration += elapsedTime.Round(time.Second)
				fmt.Printf("%d %s %s\n", entry.ID, elapsedTime.Round(time.Second).String(), entry.Name)
			} else {
				duration := time.Duration(seconds) * time.Second
				totalDuration += duration
				fmt.Printf("%d %s %s\n", entry.ID, duration.String(), entry.Name)
			}
		}
		fmt.Printf("%sTotal: %s%s\n", "\033[1m", totalDuration, "\033[0m")
	},
}

func init() {
	current := time.Now().Local()
	year, month, day := current.Date()
	rootCmd.AddCommand(entriesCmd)
	entriesCmd.Flags().StringVarP(&entriesFromFlag, "from", "", fmt.Sprintf("%d-%02d-%02d", year, month, day), fmt.Sprintf("From date format: %d-%02d-%02d", year, month, day))
}
