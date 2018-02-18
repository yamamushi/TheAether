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
	rooms	 *Rooms

}

func (h* RoomsHandler) InitRooms(s *discordgo.Session, channelID string) (err error){

	fmt.Println("Running Base Room Initialization")
	guildID, err := getGuildID(s, channelID)
	if err != nil {
		return err
	}

	h.rooms = new(Rooms)
	h.rooms.db = h.db

	h.RegisterCommands()

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
		}
	}


	// The default Welcome Channel -> To be setup correctly a server NEEDS this channel and name as the default channel
	welcomeChannelID, err := getGuildChannelIDByName(s, guildID, "welcome")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, welcomeChannelID, guildID, "Lobby")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}

	everyoneID, err := getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}
	denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{})
	alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{READ_MESSAGE_HISTORY:true, VIEW_CHANNEL: true, SEND_MESSAGES: true})
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

	err = h.CreateManagementRooms(guildID, s)
	if err != nil {
		return err
	}

	return nil
}


// CreateManagementRooms function - Useful for creating default management roles and rooms for new guilds
func (h *RoomsHandler) CreateManagementRooms(guildID string, s *discordgo.Session) (err error){

	// Create default developers role
	developerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Developer", guildID, false, false, 0, developerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default admin role
	adminsperms := h.perm.CreatePermissionInt(RolePermissions{ADMINISTRATOR:true})
	_, err = h.perm.CreateRole("Admin", guildID, false, false, 0, adminsperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default builder role
	builderperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Builder", guildID, false, false, 0, builderperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default moderator role
	moderatorperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Moderator", guildID, false, false, 0, moderatorperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default writer role
	writerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Writer", guildID, false, false, 0, writerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Developers
	_, err = h.AddRoom(s, "developers", guildID, "Management")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	developerChannelID, err := getGuildChannelIDByName(s, guildID, "developers")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, developerChannelID, guildID, "Management")
	if err != nil {
		return err
	}
	everyoneID, err := getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}
	denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{})
	err = s.ChannelPermissionSet( developerChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	developerRoleID, err := getRoleIDByName(s, guildID, "Developer")
	if err != nil {
		return err
	}
	denydevperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowdevperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( developerChannelID, developerRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}


	// Admin
	_, err = h.AddRoom(s, "admins", guildID, "Management")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	adminChannelID, err := getGuildChannelIDByName(s, guildID, "developers")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, adminChannelID, guildID, "Management")
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( adminChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	adminRoleID, err := getRoleIDByName(s, guildID, "Admin")
	if err != nil {
		return err
	}
	denyadminperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowadminperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( adminChannelID, adminRoleID, "role", allowadminperms, denyadminperms)
	if err != nil {
		return err
	}


	// Builder
	_, err = h.AddRoom(s, "builders", guildID, "Management")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	builderhannelID, err := getGuildChannelIDByName(s, guildID, "builders")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, builderhannelID, guildID, "Management")
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( builderhannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	builderRoleID, err := getRoleIDByName(s, guildID, "Builder")
	if err != nil {
		return err
	}
	denybuilderperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowbuilderperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( builderhannelID, builderRoleID, "role", allowbuilderperms, denybuilderperms)
	if err != nil {
		return err
	}


	// Moderator
	_, err = h.AddRoom(s, "moderators", guildID, "Management")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	moderatorhannelID, err := getGuildChannelIDByName(s, guildID, "moderators")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, moderatorhannelID, guildID, "Management")
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( moderatorhannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	moderatorRoleID, err := getRoleIDByName(s, guildID, "Moderator")
	if err != nil {
		return err
	}
	denymoderatorperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowmoderatorperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( moderatorhannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}


	// Writer
	_, err = h.AddRoom(s, "writers", guildID, "Management")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	writerchannelID, err := getGuildChannelIDByName(s, guildID, "writers")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, writerchannelID, guildID, "Management")
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( writerchannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	writerRoleID, err := getRoleIDByName(s, guildID, "Writer")
	if err != nil {
		return err
	}
	denywriterperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowwriterperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	err = s.ChannelPermissionSet( writerchannelID, writerRoleID, "role", allowwriterperms, denywriterperms)
	if err != nil {
		return err
	}

	return nil
}



func (h *RoomsHandler) CreateRoom(s *discordgo.Session, name string, guildID string, parentname string) {



}


func (h *RoomsHandler) AddRoom(s *discordgo.Session, name string, guildID string, parentname string) (createdroom *discordgo.Channel, err error) {

	existingrecord, err := h.rooms.GetRoomByName(name)
	if err != nil {

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


		existingrecord = Room{ID: createdchannel.ID, GuildID: guildID, Name: name, ParentID: parentID, ParentName: parentname}

		h.rooms.SaveRoomToDB(existingrecord)
		return createdroom, nil
	} else {
		return createdroom, errors.New("Room already exists in database!")
	}
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

	existingrecord, err := h.rooms.GetRoomByID(channelID)
	if err != nil {
		return err
	}
	existingrecord.ParentID = parentID
	existingrecord.ParentName = parentname
	err = h.rooms.SaveRoomToDB(existingrecord)
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
		h.ViewRoom(command[2], s, m)
		return
	}

}


func (h *RoomsHandler) ViewRoom(roomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving roomID: " + roomID)
		return
	}

	fmt.Println("Viewing Room")
	output := "\n```\n"
	output = output + "ID: " +  room.ID + "\n"
	output = output + "Name: " + room.Name + "\n"
	output = output + "Guild: " + room.GuildID + "\n"
	output = output + "ParentID: " + room.ParentID + "\n"
	output = output + "ParentName: " + room.ParentName + "\n"
	output = output + "RoleID: " + room.RoleID + "\n\n"
	output = output + "Description: " + room.Description + "\n\n"


	if room.UpID != "" {
		output = output + "Up Room: " + room.UpID + "\n"
	}
	if len(room.UpItemID) > 0 {
		itemoutput := ""
		for _, item := range room.UpItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.DownID != "" {
		output = output + "Up Room: " + room.DownID + "\n"
	}
	if len(room.DownItemID) > 0 {
		itemoutput := ""
		for _, item := range room.DownItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthID != "" {
		output = output + "Up Room: " + room.NorthID + "\n"
	}
	if len(room.NorthItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthEastID != "" {
		output = output + "Up Room: " + room.NorthEastID + "\n"
	}
	if len(room.NorthEastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthEastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.EastID != "" {
		output = output + "Up Room: " + room.EastID + "\n"
	}
	if len(room.EastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.EastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthEastID != "" {
		output = output + "Up Room: " + room.SouthEastID + "\n"
	}
	if len(room.SouthEastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthEastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthID != "" {
		output = output + "Up Room: " + room.SouthID + "\n"
	}
	if len(room.SouthItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthWestID != "" {
		output = output + "Up Room: " + room.SouthWestID + "\n"
	}
	if len(room.SouthWestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthWestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.WestID != "" {
		output = output + "Up Room: " + room.WestID + "\n"
	}
	if len(room.WestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.WestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthWestID != "" {
		output = output + "Up Room: " + room.NorthWestID + "\n"
	}
	if len(room.NorthWestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthWestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	output = output + "\n```\n"

	s.ChannelMessageSend(m.ChannelID, "Room Details: " + output)
	return
}



