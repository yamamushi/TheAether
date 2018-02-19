package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"errors"
)

type RegistrationHandler struct {

	callback *CallbackHandler
	conf     *Config
	db       *DBHandler
	perm     *PermissionsHandler
	registry *CommandRegistry
	dg       *discordgo.Session
	user     *UserHandler
	ch       *ChannelHandler
	rooms	 *Rooms
	guilds 	 *GuildsManager

}


func (h *RegistrationHandler) Init() {

	h.RegisterCommands()

}


func (h *RegistrationHandler) RegisterCommands() (err error) {

	h.registry.Register("register", "Register a new account", "")
	h.registry.AddGroup("register", "player")
	h.registry.AddChannel("register", h.conf.MainConfig.LobbyChannelID)

	return nil

}



// Read function
func (h *RegistrationHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	// This should register all new users, presumably we want this done here because this is the first
	// command a user should have access to.
	h.user.CheckUser(m.Author.ID, s, m.ChannelID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"register") {
		if user.Registered != "" {
			//_, payload := CleanCommand(m.Content, h.conf)

			if user.CheckRole("player") {
				h.StartRegistration(s, m)
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "You are already registered! If you continue to have issues please ask an Admin for assistance.")
			return
		}
	}
	if strings.HasPrefix(m.Content, cp+"roll-abilities") {
		if user.Registered != "" {
			//_, payload := CleanCommand(m.Content, h.conf)

			if user.CheckRole("player") {
				h.StartRegistration(s, m)
			}

		}
	}
}



// ParseCommand function
func (h *RegistrationHandler) StartRegistration(s *discordgo.Session, m *discordgo.MessageCreate) {
/*
	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}
*/
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error finding user: " + err.Error())
		return
	}


	privateMessage := ":sunrise_over_mountains: Welcome to The Aether! ```\n"
	privateMessage = privateMessage + "A basic avatar has been summoned for you, however it is "


	userprivatechannel, err := s.UserChannelCreate(user.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error starting Registration: " + err.Error())
		return
	}

	err = h.SetRegistrationStatus("attributes", user.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error starting Registration: " + err.Error())
		return
	}

	s.ChannelMessageSend(userprivatechannel.ID, privateMessage)
	return
}


func (h *RegistrationHandler) SetRegistrationStatus(status string, userID string) (err error){

	switch status {
		case "attributes":
			break;
		case "complete":
			break;
		default:
			return errors.New("Invalid registration status update")
	}

	user, err := h.db.GetUser(userID)
	if err != nil {
		return err
	}

	user.RegistrationStatus = status
	err = h.user.user.SaveUserToDB(user)
	if err != nil {
		return err
	}

	return nil
}


