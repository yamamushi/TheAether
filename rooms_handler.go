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


	// Create default registered user role
	registeredperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("registered", guildID, false, false, 16777215, registeredperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}

	// Create default crossroads location role
	crossroadsperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Crossroads", guildID, false, false, 0, crossroadsperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}	}


	// The default Welcome Channel -> To be setup correctly a server NEEDS this channel and name as the default channel
	welcomeChannelID, err := getGuildChannelIDByName(s, guildID, "welcome")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, welcomeChannelID, guildID, "Lobby")
	if err != nil {
		return err
	}

	everyoneID, err := getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}
	denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{})
	alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, SEND_MESSAGES: true})
	err = s.ChannelPermissionSet( welcomeChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}

	registeredroleID, err := getRoleIDByName(s, guildID, "registered")
	if err != nil {
		return err
	}
	denyregisteredperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	allowregisteredperms := h.perm.CreatePermissionInt(RolePermissions{})
	err = s.ChannelPermissionSet( welcomeChannelID, registeredroleID, "role", allowregisteredperms, denyregisteredperms)
	if err != nil {
		return err
	}


	// Crossroads
	_, err = h.AddRoom(s, "crossroads", guildID, "The Aether")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	crossroadsChannelID, err := getGuildChannelIDByName(s, guildID, "crossroads")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, crossroadsChannelID, guildID, "The Aether")
	if err != nil {
		return err
	}

	everyoneID, err = getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}
	denyeveryoneperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	alloweveryoneperms = h.perm.CreatePermissionInt(RolePermissions{})
	err = s.ChannelPermissionSet( crossroadsChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}

	crossroadsRoleID, err := getRoleIDByName(s, guildID, "Crossroads")
	if err != nil {
		return err
	}
	denyrossroadsperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowcrossroadsperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( crossroadsChannelID, crossroadsRoleID, "role", allowcrossroadsperms, denyrossroadsperms)
	if err != nil {
		return err
	}
	return nil
}



func (h *RoomsHandler) CreateRoom(s *discordgo.Session, name string, guildID string, parentname string) {



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
					return createdroom, errors.New("Channel with parent category already exists - Channel:" + name + " Parent: " + parentname)
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


func (h *RoomsHandler) MoveRoom(s *discordgo.Session, channelID string, guildID string, parentname string) (err error) {

	if parentname == "" {
		return errors.New("No parent category supplied")
	}

	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return err
	}

	parentID := ""

	for _, channel := range channels {
		if channel.Name == parentname {
			parentID = channel.ID
		}
	}
	if parentID == "" {
		return errors.New("No parent category supplied")
	}

	modifyChannel := discordgo.ChannelEdit{ParentID: parentID}
	_, err = s.ChannelEditComplex(channelID, &modifyChannel)
	if err != nil {
		return err
	}

	return nil
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
	if command[1] == "view" {
		if len(command) == 2 {
			s.ChannelMessageSend(m.ChannelID, "view requires a room argument")
			return
		}
		h.ViewRoom(command[3], s, m)
		return
	}

}


func (h *RoomsHandler) ViewRoom(roomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	//roomID = CleanChannel(room)



}



