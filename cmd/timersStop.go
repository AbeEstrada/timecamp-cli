package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var StopTimerID string

var timersStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a timer",
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := os.Getenv("TIMECAMP_API_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Missing TIMECAMP_API_TOKEN environment variable")
			return
		}

		url := "https://app.timecamp.com/third_party/api/timer"

		payload := strings.NewReader("{\"action\":\"stop\"}")
		if StopTimerID != "" {
			type Payload struct {
				Action string `json:"action"`
				TaskID string `json:"task_id"`
			}
			payloadStruct := Payload{
				Action: "stop",
				TaskID: StopTimerID,
			}
			payloadJSON, err := json.Marshal(payloadStruct)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			payload = strings.NewReader(string(payloadJSON))
			StopTimerID = ""
		}

		req, _ := http.NewRequest("POST", url, payload)

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
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
			Elapsed   int    `json:"elapsed"`
			EntryID   string `json:"entry_id"`
			EntryTime int    `json:"entry_time"`
		}

		var entry Response
		err := json.Unmarshal([]byte(body), &entry)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		elapsed := time.Duration(entry.Elapsed) * time.Second
		fmt.Printf("Stopped timer: %s\n", elapsed.String())
	},
}

func init() {
	timersCmd.AddCommand(timersStopCmd)
	timersStopCmd.Flags().StringVarP(&StopTimerID, "id", "i", "", "Timer ID (required)")
}
