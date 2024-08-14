package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Groups struct {
	Groups []DeployGroup `json:"groups"`
}

// DeployGroup represents a group of deployments.
type DeployGroup struct {
	GroupName    string   `json:"group_name"`    // The name of the deployment group.
	Apps         []string `json:"apps"`          // The list of applications in the deployment group.
	SlackChannel string   `json:"slack_channel"` // The Slack channel associated with the deployment group.
}

func NewGroups() *Groups {
	return &Groups{}
}

func (g *Groups) ReadGroups(input string) error {
	// Unmarshal the JSON data into the Groups struct.
	err := json.Unmarshal([]byte(input), g)
	if err != nil {
		return fmt.Errorf("failed to unmarshal deployment groups: %w", err)
	}

	return nil
}

// Iterate iterates over the deployment groups and calls the specified function for each group.
func (g *Groups) CheckContainerImage(fn func(DeployGroup, map[string]string), deployRolloutDetails map[string]string) {
	for _, group := range g.Groups {
		fn(group, deployRolloutDetails)
	}
}

func FindMismatch(group DeployGroup, deployRolloutDetails map[string]string) {
	redisValues := HGETALL(group.GroupName)
	if len(redisValues) == 0 {
		redisValues[IMAGE_STATUS] = IMAGE_STATUS_ALL_SAME
		redisValues[NOTIFIED_AT] = "-1"
	}

	mismatchDetected := false
	image := deployRolloutDetails[group.Apps[0]]
	if !strings.Contains(redisValues[CLUSTERS], clusterName) {
		if len(redisValues[CLUSTERS]) != 0 {
			redisValues[CLUSTERS] = redisValues[CLUSTERS] + "," + clusterName
		} else {
			redisValues[CLUSTERS] = clusterName
		}
	}
	redisValues[clusterName] = image
	sugar.Debug("Checking group: ", group.GroupName)
	for _, app := range group.Apps {
		if image == "" {
			image = deployRolloutDetails[app]
			redisValues[clusterName] = image
		}
		sugar.Debugf("Checking app: %s, Image: %s", app, deployRolloutDetails[app])
		appImage, ok := deployRolloutDetails[app]
		if !ok {
			sugar.Warnf("App: %s not found in deployment/rollout", app)
		}
		if image != appImage {
			mismatchDetected = true
			sugar.Warnf("Mismatch found in group: %s, Apps: %s", group.GroupName, group.Apps)
			statusHistory := redisValues[IMAGE_STATUS]
			if statusHistory == IMAGE_STATUS_DIFF_APPS {
				if redisValues[NOTIFIED_AT] == "-1" || isMoreThanThreeHours(redisValues[NOTIFIED_AT]) {
					redisValues[NOTIFIED_AT] = time.Now().UTC().Format("2006-01-02 15:04")
					SendSlackMessage(SLACK_CHANNEL, group.GroupName, fmt.Sprintf("*Cluster(s):* %s\n*Apps:* %s", clusterName, group.Apps))
				} else {
					if redisValues[NOTIFIED_AT] == "-1" {
						sugar.Warnf("Will notify in next cycle if required since images were in sync before.")
					} else {
						sugar.Warnf("Already notified in last 3 hours, will not notify again.")
					}
				}
			} else {
				sugar.Warnf("Will notify in next cycle if required since images were in sync before.")
				redisValues[IMAGE_STATUS] = IMAGE_STATUS_DIFF_APPS
			}
			HSETAll(group.GroupName, redisValues)
			break
		}
	}
	if !mismatchDetected {
		clusters := strings.Split(redisValues[CLUSTERS], ",")
		if len(clusters) > 1 {
			clusterImage := redisValues[clusters[0]]
			for _, cluster := range clusters {
				if clusterImage != redisValues[cluster] {
					mismatchDetected = true
					sugar.Warnf("Mismatch found for group: %s, Apps: %s, across clusters %s", group.GroupName, group.Apps, redisValues[CLUSTERS])
					statusHistory := redisValues[IMAGE_STATUS]
					if statusHistory == IMAGE_STATUS_DIFF_CLUSTERS {
						if redisValues[NOTIFIED_AT] == "-1" || isMoreThanThreeHours(redisValues[NOTIFIED_AT]) {
							redisValues[NOTIFIED_AT] = time.Now().UTC().Format("2006-01-02 15:04")
							SendSlackMessage(SLACK_CHANNEL, group.GroupName, fmt.Sprintf("*Images deployed across clusters - %s are different.*\n*Apps:* %s", redisValues[CLUSTERS], group.Apps))
						} else {
							if redisValues[NOTIFIED_AT] == "-1" {
								sugar.Warnf("Will notify in next cycle if required since images were in sync before.")
							} else {
								sugar.Warnf("Already notified in last 3 hours, will not notify again.")
							}
						}
					} else {
						sugar.Warnf("Will notify in next cycle if required since images were in sync before.")
						if statusHistory == IMAGE_STATUS_ALL_SAME {
							redisValues[IMAGE_STATUS] = IMAGE_STATUS_DIFF_CLUSTERS
						}
					}
					break
				}
			}
		}
		if !mismatchDetected {
			redisValues[IMAGE_STATUS] = IMAGE_STATUS_ALL_SAME
			redisValues[NOTIFIED_AT] = "-1"
			sugar.Infof("No mismatch found in group: %s", group.GroupName)
		}
		HSETAll(group.GroupName, redisValues)
	}
}

func isMoreThanThreeHours(notifiedAt string) bool {
	notifiedAtTime, _ := time.Parse("2006-01-02 15:04", notifiedAt)
	return time.Since(notifiedAtTime).Hours() > 3
}
