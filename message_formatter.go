package main

import "strings"

// FormatEventMessage function
func FormatEventMessage(message string, userID string, channelID string) (formatted string) {
	formatted = strings.Replace(message, "_user_", "<@"+userID+">", -1)
	formatted = strings.Replace(formatted, "_room_", "<#"+channelID+">", -1)
	return formatted
}
