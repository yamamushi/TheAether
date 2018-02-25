package main

import (

	"github.com/bwmarrin/discordgo"
	"strings"
)

// WelcomeHandler struct
type WelcomeHandler struct {

	conf 	*Config
	user 	*UserHandler
	db 		*DBHandler

}


// Read function
func (h *WelcomeHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	if strings.HasPrefix(m.Content, cp+"about") {

		aboutmessage := ":bulb: The Aether :bulb: ```\n"
		aboutmessage = aboutmessage + "The Aether is a roleplaying game for Discord. Within The Aether you may take many roles.\n\n"
		aboutmessage = aboutmessage + "Will you become a traveled adventurer or a rich king? Perhaps a ship merchantman or a shopkeeper?\n\n"
		aboutmessage = aboutmessage + "Whatever you choose to become, The Aether welcomes you on your journey!"
		aboutmessage = aboutmessage + "\n```\n"

		s.ChannelMessageSend(m.ChannelID, aboutmessage)
		return
	}

}

// ReadNewMember function
func (h *WelcomeHandler) ReadNewMember(s *discordgo.Session, m *discordgo.GuildMemberAdd) {

	// We do not want to send a welcome message to anyone except on first join in the central (main) guild
	if m.GuildID != h.conf.MainConfig.CentralGuildID {
		return
	}

	welcomemessage := ""
	welcomemessage = welcomemessage + "Welcome to The Aether " + m.Member.User.Mention() + "!\n\n"
	welcomemessage = welcomemessage + "Please take a moment to read the #serverrules before proceeding. This is a roleplay enforced "
	welcomemessage = welcomemessage + "game and it is pivotal to everyones enjoyment of it that you stay in character within the "
	welcomemessage = welcomemessage + "applicable channels. Please use ooc channels for out of character chat.\n\n"
	welcomemessage = welcomemessage + "When you are ready, you may run ~register and follow the instructions in the private message you receive. "
	welcomemessage = welcomemessage + "For more information about The Aether you can use the ~about command."
	welcomemessage = welcomemessage + "\n"

	s.ChannelMessageSend(h.conf.MainConfig.LobbyChannelID, welcomemessage)
	return
}