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


/*
	guildChannels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		fmt.Print("Error: Could not retrieve guild channels in welcome handler read!")
		return
	}

	for _, channel := range guildChannels {
		if channel.Name == "welcome" {

		}
	}
*/

	s.ChannelMessageSend(h.conf.MainConfig.LobbyChannelID, "Welcome to The Aether " + m.Member.User.Mention() + "!")

	return
}