package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"errors"
)

type RoomsHandler struct {

	callback *CallbackHandler
	conf     *Config
	db       *DBHandler
	perm     *PermissionsHandler
	registry *CommandRegistry
	dg       *discordgo.Session
	user     *UserHandler
	ch       *ChannelHandler

}

func (h* RoomsHandler) InitRooms(s *discordgo.Session, channelID string) (err error){

	fmt.Println("Running Base Room Initialization")
	guildID, err := getGuildID(s, channelID)
	if err != nil {
		return err
	}

	perms := h.perm.CreatePermissionInt(RolePermissions{MENTION_EVERYONE: true})
	_, err = h.perm.CreateRole("testtwo", guildID, true, true, 100, perms, s)
	if err != nil {
		//return err
	}

	_, err = h.AddRoom(s, "test", guildID, "Management")
	if err != nil {
		//return err
	}

	/*
	overwrite, err := h.perm.CreatePermissionOverwrite(createdrole.ID, "MENTION_EVERYONE", false)
	if err != nil {
		return err
	}
	*/

	denyperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	allowperms := h.perm.CreatePermissionInt(RolePermissions{})

	everyoneID, err := getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}
	fmt.Println("Everyone: " + everyoneID)

	testChannelID, err := getGuildChannelIDByName(s, guildID, "test")
	if err != nil {
		return err
	}

	err = s.ChannelPermissionSet( testChannelID, everyoneID, "role", allowperms, denyperms)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) AddRoom(s *discordgo.Session, name string, guildID string, parentname string) (createdroom *discordgo.Channel, err error) {

	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return createdroom, err
	}

	parentID := ""
	for _, channel := range channels {
		if channel.Name == name {
			if channel.ParentID != "" {
				parentchannel, err := s.Channel(channel.ParentID)
				if err != nil {
					return createdroom, err
				}

				if parentchannel.Name == parentname {
					return createdroom, errors.New("Channel with parent category already exists: " + name + " " + parentname)
				}
			}
		}

		if channel.Name == parentname {
			parentID = channel.ID
		}
	}

	if parentID == "" {
		return createdroom, errors.New("Parent channel not found: " + parentname)
	}

	createdchannel, err := s.GuildChannelCreate(guildID, name, "text")
	if err != nil {
		return createdroom, err
	}

	modifyChannel := discordgo.ChannelEdit{Name: createdchannel.Name, ParentID: parentID}

	createdroom, err = s.ChannelEditComplex(createdchannel.ID, &modifyChannel)
	if err != nil {
		return createdroom, err
	}

	return createdroom, nil
}




// RegisterCommands function
func (h *RoomsHandler) RegisterCommands() (err error) {

	h.registry.Register("room", "Manage rooms for this server", "")
	return nil

}


// Read function
func (h *RoomsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println("Error finding user")
		return
	}
	if !user.CheckRole("builder") {
		return
	}

	if strings.HasPrefix(m.Content, cp+"room") {
		if h.registry.CheckPermission("room", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.CheckRole("moderator") {
				h.ParseCommand(command, s, m)
			}
		}
	}
}


// ParseCommand function
func (h *RoomsHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(command) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'room' command, see usage for more info")
		return
	}

}


