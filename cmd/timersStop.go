package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

		fmt.Println(string(body))
	},
}

func init() {
	timersCmd.AddCommand(timersStopCmd)
	timersStopCmd.Flags().StringVarP(&StopTimerID, "id", "i", "", "Timer ID (required)")
}
