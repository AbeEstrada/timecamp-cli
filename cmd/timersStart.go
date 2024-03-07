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

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer",
	Run: func(cmd *cobra.Command, args []string) {
		apiToken := os.Getenv("TIMECAMP_API_TOKEN")
		if apiToken == "" {
			fmt.Println("Error: Missing TIMECAMP_API_TOKEN environment variable")
			return
		}

		url := "https://app.timecamp.com/third_party/api/timer"

		payload := strings.NewReader("{\"action\":\"start\"}")

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
			EntryID int `json:"entry_id"`
		}

		var parsedData Response
		err := json.Unmarshal(body, &parsedData)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		entryID := parsedData.EntryID
		fmt.Println("Entry ID:", entryID)
	},
}

func init() {
	timersCmd.AddCommand(startCmd)
}
