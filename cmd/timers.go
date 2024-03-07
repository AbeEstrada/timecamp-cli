package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "Get information about running timers",
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := os.Getenv("TIMECAMP_API_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Missing TIMECAMP_API_TOKEN environment variable")
			return
		}

		url := "https://app.timecamp.com/third_party/api/timer_running"

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
			TimerID   string  `json:"timer_id"`
			UserID    string  `json:"user_id"`
			TaskID    *string `json:"task_id"` // Nullable field
			StartedAt string  `json:"started_at"`
			Name      *string `json:"name"` // Nullable field
		}

		var timers []Response
		err := json.Unmarshal([]byte(body), &timers)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		loc, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var totalDuration time.Duration
		for _, timer := range timers {
			fmt.Printf("Timer ID: %s\n", timer.TimerID)
			fmt.Printf("Started At: %s\n", timer.StartedAt)
			givenTime, err := time.ParseInLocation("2006-01-02 15:04:05", timer.StartedAt, loc)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			elapsedTime := time.Since(givenTime)
			totalDuration += elapsedTime.Round(time.Second)
			fmt.Println(elapsedTime.Round(time.Second).String())

			if len(timers) > 1 {
				fmt.Println("---")
			}
		}

		fmt.Printf("%sTotal: %s%s\n", "\033[1m", totalDuration.String(), "\033[0m")
	},
}

func init() {
	rootCmd.AddCommand(timersCmd)
}
