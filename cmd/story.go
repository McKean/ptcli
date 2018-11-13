// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Label defines a label
type Label struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"project_id"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Story defines a story
type Story struct {
	Kind          string    `json:"kind"`
	ID            int       `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	AcceptedAt    time.Time `json:"accepted_at"`
	Estimate      int       `json:"estimate"`
	StoryType     string    `json:"story_type"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	CurrentState  string    `json:"current_state"`
	RequestedByID int       `json:"requested_by_id"`
	ExternalID    string    `json:"external_id"`
	IntegrationID int       `json:"integration_id"`
	URL           string    `json:"url"`
	ProjectID     int       `json:"project_id"`
	OwnerIDs      []int     `json:"owner_ids"`
	Labels        []Label   `json:"labels"`
	OwnedByID     int       `json:"owned_by_id"`
}

// Iteration is the struct of the iteration
type Iteration struct {
	Number       int       `json:"number"`
	ProjectID    int       `json:"project_id"`
	Length       int       `json:"length,omitempty"`
	TeamStrength int       `json:"team_strength,omitempty"`
	Stories      []Story   `json:"stories"`
	Start        time.Time `json:"start,omitempty"`
	Finish       time.Time `json:"finish,omitempty"`
	Kind         string    `json:"kind,omitempty"`
}

// storyCmd represents the story command
var storyCmd = &cobra.Command{
	Use:   "story",
	Short: "Add stories to your Pivotal Tracker project",
	Long: `You can add new stories and set if they should be a
	feature, bug (-b), or chore (-c).

	By default a story is a feature and is added to the ice box.
	If -i (important/immediate) flag is added it's added to the top of
	the Backlog.

	Labels can be added with the -l flag. Add multiple by separating them
	with "," (no spaces).`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Provide story title")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var storyType string
		var firstBacklogStory int = 0
		var currentState = "unscheduled"
		client := &http.Client{}

		isBug, _ := cmd.Flags().GetBool("bug")
		isChore, _ := cmd.Flags().GetBool("chore")
		token, _ := cmd.Flags().GetString("token")
		project, _ := cmd.Flags().GetString("project")
		labels, _ := cmd.Flags().GetString("labels")
		labelArray := strings.Split(labels, ",")
		important, _ := cmd.Flags().GetBool("important")

		if important {
			req, err := http.NewRequest("GET", fmt.Sprintf("https://www.pivotaltracker.com/services/v5/projects/%s/iterations?limit=1&offset=1&scope=current", project), nil)
			req.Header.Add("X-TrackerToken", token)
			req.Header.Add("Content-Type", "application/json")

			response, err := client.Do(req)

			if err != nil {
				fmt.Printf("The HTTP request failed with error %s\n", err)
			} else {
				data, _ := ioutil.ReadAll(response.Body)
				var iteration []Iteration
				if err := json.Unmarshal(data, &iteration); err != nil {
					fmt.Println(err)
				}
				firstBacklogStory = iteration[0].Stories[0].ID
			}
		}

		if firstBacklogStory > 0 {
			currentState = "unstarted"
		} else {
			firstBacklogStory = 0
		}

		if len(labelArray[0]) == 0 {
			labelArray = nil
		}

		if isChore == true && isBug == true {
			fmt.Println("Story type cannot be chore and bug, both will be ignored")
		} else if isBug {
			storyType = "bug"
		} else if isChore {
			storyType = "chore"
		} else {
			storyType = "feature"
		}

		estimate, _ := cmd.Flags().GetInt("estimate")

		type Params struct {
			Name         string   `json:"name"`
			StoryType    string   `json:"story_type"`
			Labels       []string `json:"labels,omitempty"`
			CurrentState string   `json:"current_state,omitempty"`
			AfterID      int      `json:"before_id,omitempty"`
			Estimate     int      `json:"estimate,omitempty"`
		}

		p := Params{
			Name:         strings.Join(args, " "),
			StoryType:    storyType,
			Labels:       labelArray,
			CurrentState: currentState,
			Estimate:     estimate,
			AfterID:      firstBacklogStory,
		}

		jsonValue, err := json.Marshal(p)

		req, err := http.NewRequest("POST", fmt.Sprintf("https://www.pivotaltracker.com/services/v5/projects/%s/stories", project), bytes.NewBuffer(jsonValue))
		req.Header.Add("X-TrackerToken", token)
		req.Header.Add("Content-Type", "application/json")

		response, err := client.Do(req)

		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			var story Story
			if err := json.Unmarshal(data, &story); err != nil {
				fmt.Println(err)
			}
			fmt.Println("Story created: " + strconv.Itoa(story.ID))
		}
	},
}

func init() {
	rootCmd.AddCommand(storyCmd)
	storyCmd.Flags().IntP("estimate", "e", 0, "Set story estimate points")
	storyCmd.Flags().BoolP("bug", "b", false, "Set story type as bug")
	storyCmd.Flags().BoolP("chore", "c", false, "Set story type as chore")
	storyCmd.Flags().StringP("labels", "l", "", "Set the labels")
	viper.AutomaticEnv()
	storyCmd.Flags().StringP("token", "t", viper.GetString("PIVOTAL_TOKEN"), "Set Pivotal token")
	storyCmd.Flags().StringP("project", "p", viper.GetString("PIVOTAL_PROJECT"), "Set Pivotal project")
	storyCmd.Flags().BoolP("important", "i", false, "Add immediately (important) to top of backlog")
}
