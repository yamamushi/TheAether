package main

import (

	"github.com/bwmarrin/discordgo"

	)

type WelcomeHandler struct {

	conf 	*Config
	user 	*UserHandler
	db 		*DBHandler

}



func (h *WelcomeHandler) Read(s *discordgo.Session, m *discordgo.GuildMemberAdd) {

	// We do not want to send a welcome message to anyone except on first join in the central (main) guild
	if m.GuildID != h.conf.MainConfig.CentralGuildID {
		return
	}

	welcomemessage := ""
	welcomemessage = welcomemessage + "Welcome to The Aether " + m.Member.User.Mention() + "!\n\n"
	welcomemessage = welcomemessage + "Please take a moment to read the #serverrules before proceeding. This is a roleplay enforced "
	welcomemessage = welcomemessage + "game and it is pivotal to everyones enjoyment of it that you stay in character within the "
	welcomemessage = welcomemessage + "applicable channels. Please use ooc channels for out of character chat. "
	welcomemessage = welcomemessage + "```\n"

		s.ChannelMessageSend(h.conf.MainConfig.LobbyChannelID, welcomemessage)

	return
}