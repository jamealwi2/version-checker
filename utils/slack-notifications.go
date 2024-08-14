package utils

import "github.com/slack-go/slack"

func SendSlackMessage(channel, group, message string) {
	slackClient := slack.New("<TOKEN>")

	attachment := slack.Attachment{
		Pretext: "",
		Text:    message,
		Color:   "#ffff00",
	}

	// Construct the message options
	options := []slack.MsgOption{
		slack.MsgOptionText("*Container image mismatch for _"+group+"_ deploys*", false),
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionUsername(BOT_NAME),
		slack.MsgOptionIconURL(BOT_AVATAR),
	}

	// Send the message
	_, _, err := slackClient.PostMessage(channel, options...)
	if err != nil {
		sugar.Errorf("failed to send message: %v", err)
	}
}
