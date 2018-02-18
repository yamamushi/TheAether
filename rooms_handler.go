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
	_, err = h.AddRoom(s, "crossroads", guildID, "The Aether", "")
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

	room, err := h.rooms.GetRoomByID(crossroadsChannelID)
	if err != nil {
		return err
	}

	updatecrossroads := true
	for _, roleid := range room.RoleIDs {
		if roleid == crossroadsRoleID {
			updatecrossroads = false
		}
	}
	if updatecrossroads{
		room.RoleIDs = append(room.RoleIDs, crossroadsRoleID)

		err = h.rooms.SaveRoomToDB(room)
		if err != nil {
			return err
		}
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
	_, err = h.AddRoom(s, "developers", guildID, "Management", "")
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
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
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
	_, err = h.AddRoom(s, "admins", guildID, "Management", "")
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
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
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
	_, err = h.AddRoom(s, "builders", guildID, "Management", "")
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
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
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
	_, err = h.AddRoom(s, "moderators", guildID, "Management", "")
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
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
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
	_, err = h.AddRoom(s, "writers", guildID, "Management", "")
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
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
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

func (h *RoomsHandler) DeleteRoom(s *discordgo.Session, name string, guildID string, parentname string) {



}



func (h *RoomsHandler) AddRoom(s *discordgo.Session, name string, guildID string, parentname string, transferID string) (createdroom *discordgo.Channel, err error) {

	existingrecord, err := h.rooms.GetRoomByName(name, guildID)
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

		err = h.registry.AddChannel("travel", createdroom.ID)
		if err != nil {
			return createdroom, err
		}

		existingrecord = Room{ID: createdchannel.ID, GuildID: guildID, Name: name, ParentID: parentID, ParentName: parentname,
		TransferID: transferID}

		everyoneID, err := getGuildEveryoneRoleID(s, guildID)
		if err != nil {
			return createdroom, err
		}
		denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
		alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{})
		err = s.ChannelPermissionSet( createdroom.ID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
		if err != nil {
			return createdroom, err
		}

		h.rooms.SaveRoomToDB(existingrecord)
		return createdroom, nil
	} else {
		return createdroom, errors.New("Room already exists in database!")
	}
}


func (h *RoomsHandler) RemoveRoom(s *discordgo.Session, name string, guildID string) (err error) {

	existingrecord, err := h.rooms.GetRoomByName(name, guildID)
	if err != nil{
		return err
	}

	s.ChannelDelete(existingrecord.ID)

	err = h.registry.RemoveChannel("travel", existingrecord.ID)
	if err != nil {
		return err
	}

	return h.rooms.RemoveRoomFromDB(existingrecord)
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

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

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
	if command[1] == "add" {
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "add requires at least one argument: <name> <transferID>")
			return
		}

		parentname := "The Aether"
		transferID := ""
		if len(command) > 3 {
			transferID = command[3]
		}

		channel, err := h.AddRoom(s, command[2], guildID, parentname, transferID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error adding channel: " + err.Error())
			return
		}


		formatted, err := h.FormatRoomInfo(channel.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error Retrieving Room: " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Channel Created: " + formatted)
		return
	}
	if command[1] == "linkdirection" {
		if len(command) < 5 {
			s.ChannelMessageSend(m.ChannelID, "linkdirection requires three arguments: <from> <to> <direction>")
			return
		}
		h.LinkDirection(command[4], command[2], command[3], s, m)
		return
	}
	if command[1] == "remove" {
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "remove requires an argument: <room name>")
			return
		}

		roomname := ""
		if strings.Contains(command[2], "#"){
			roomname := strings.TrimPrefix(command[2], "<#")
			roomname = strings.TrimSuffix(roomname, ">")

			room, err := h.rooms.GetRoomByID(roomname)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error removing channel: " + err.Error())
				return
			}
			err = h.RemoveRoom(s, room.Name, guildID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error removing channel: " + err.Error())
				return
			}
		} else {
			err := h.RemoveRoom(s, roomname, guildID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error removing channel: " + err.Error())
				return
			}
		}

		s.ChannelMessageSend(m.ChannelID, "Channel " + command[2] + " removed.")
		return
	}
	if command[1] == "linkrole" {
		if len(command) < 4 {
			s.ChannelMessageSend(m.ChannelID, "linkrole requires two arguments - <rolename> <room>")
			return
		}

		h.LinkRole(command[2], command[3], s, m)
		return
	}
	if command[1] == "setupserver" {
		if m.Author.ID != h.conf.MainConfig.ClusterOwnerID {
			s.ChannelMessageSend(m.ChannelID, "Only the cluster owner can run this command.")
			return
		}
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "setupserver requires an acknowledgement flag (y/n)")
			return
		}
		command[2] = strings.ToLower(command[2])
		if command[2] == "y" {
			err := h.SetupNewServer(s,m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not setup new server: " + err.Error())
				return
			}
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "setupserver requires an acknowledgement flag (y/n)")
			return
		}

	}
}


func (h *RoomsHandler) SetupNewServer(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	err = h.CreateManagementRooms(guildID, s)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) LinkRole(rolename string, roomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	roleID, err := getRoleIDByName(s, guildID, rolename)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving roleID: " + err.Error())
		return
	}

	roomID = CleanChannel(roomID)
	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving roomID: " + err.Error())
		return
	}

	for _, roleid := range room.RoleIDs {
		if roleid == roleID {
			s.ChannelMessageSend(m.ChannelID, "Room is already linked to role!")
			return
		}
	}
	room.RoleIDs = append(room.RoleIDs, roleID)

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error updating DB: " + err.Error())
		return
	}

	denyrperms := 0
	allowperms := 0
	if room.TransferID != "" {
		denyrperms = h.perm.CreatePermissionInt(RolePermissions{SEND_MESSAGES:false, READ_MESSAGE_HISTORY:false})
		allowperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	} else {
		denyrperms = h.perm.CreatePermissionInt(RolePermissions{})
		allowperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true})
	}

	err = s.ChannelPermissionSet( room.ID, roleID, "role", allowperms, denyrperms)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error setting permissions: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role " + rolename + " linked to " + room.Name)
	return
}

func (h *RoomsHandler) FormatRoomInfo(roomID string) (formatted string, err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return "", err
	}

	output := "\n```\n"
	output = output + "ID: " +  room.ID + "\n"
	output = output + "Name: " + room.Name + "\n"
	output = output + "Guild: " + room.GuildID + "\n"
	output = output + "TransferID: " + room.TransferID + "\n"
	output = output + "ParentID: " + room.ParentID + "\n"
	output = output + "ParentName: " + room.ParentName + "\n"

	roles := ""
	for i, role := range room.RoleIDs {
		if i > 0 {
			roles = ", " + role
		} else {
			roles = role
		}
	}
	output = output + "RoleID: " + roles + "\n\n"
	output = output + "Description: " + room.Description + "\n\n"


	if room.UpID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.UpID)
		if err != nil {
			return formatted, err
		}
		output = output + "Up Room: " + room.UpID + " - " + linkedroom.Name + " \n"
	}
	if len(room.UpItemID) > 0 {
		itemoutput := ""
		for _, item := range room.UpItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.DownID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.DownID)
		if err != nil {
			return formatted, err
		}
		output = output + "Down Room: " + room.DownID + " - " + linkedroom.Name + " \n"
	}
	if len(room.DownItemID) > 0 {
		itemoutput := ""
		for _, item := range room.DownItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.NorthID)
		if err != nil {
			return formatted, err
		}
		output = output + "North Room: " + room.NorthID + " - " + linkedroom.Name + " \n"
	}
	if len(room.NorthItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthEastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.NorthEastID)
		if err != nil {
			return formatted, err
		}
		output = output + "NorthEast Room: " + room.NorthEastID + " - " + linkedroom.Name + " \n"
	}
	if len(room.NorthEastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthEastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.EastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.EastID)
		if err != nil {
			return formatted, err
		}
		output = output + "East Room: " + room.EastID + " - " + linkedroom.Name + " \n"
	}
	if len(room.EastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.EastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthEastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.SouthEastID)
		if err != nil {
			return formatted, err
		}
		output = output + "SouthEast Room: " + room.SouthEastID + " - " + linkedroom.Name + " \n"
	}
	if len(room.SouthEastItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthEastItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.SouthID)
		if err != nil {
			return formatted, err
		}
		output = output + "South Room: " + room.SouthID + " - " + linkedroom.Name + " \n"
	}
	if len(room.SouthItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.SouthWestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.SouthWestID)
		if err != nil {
			return formatted, err
		}
		output = output + "SouthWest Room: " + room.SouthWestID + " - " + linkedroom.Name + " \n"
	}
	if len(room.SouthWestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.SouthWestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.WestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.WestID)
		if err != nil {
			return formatted, err
		}
		output = output + "West Room: " + room.WestID + " - " + linkedroom.Name + " \n"
	}
	if len(room.WestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.WestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	if room.NorthWestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(room.NorthWestID)
		if err != nil {
			return formatted, err
		}
		output = output + "NorthWest Room: " + room.NorthWestID + " - " + linkedroom.Name + " \n"
	}
	if len(room.NorthWestItemID) > 0 {
		itemoutput := ""
		for _, item := range room.NorthWestItemID {
			itemoutput = itemoutput +", " + item
		}
		output = output + "Up Room Required Items: " + itemoutput + "\n"
	}

	output = output + "\n```\n"
	return output, nil
}


func (h *RoomsHandler) ViewRoom(roomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	formatted, err := h.FormatRoomInfo(roomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error Retrieving Room: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Room Details: " + formatted)
	return
}


func (h *RoomsHandler) LinkDirection(direction string, fromroomID string, toroomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	fromroomID = CleanChannel(fromroomID)
	fromroom, err := h.rooms.GetRoomByID(fromroomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving fromroomID: " + err.Error())
		return
	}
	if fromroom.GuildID != guildID {
		s.ChannelMessageSend(m.ChannelID, "Guild ID's Do Not Match: " + err.Error())
		return
	}

	toroomID = CleanChannel(toroomID)
	toroom, err := h.rooms.GetRoomByID(toroomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving toroomID: " + err.Error())
		return
	}
	if toroom.GuildID != guildID {
		s.ChannelMessageSend(m.ChannelID, "Guild ID's Do Not Match: " + err.Error())
		return
	}

	direction = strings.ToLower(direction)

	linked, err := h.rooms.IsRoomLinkedTo(fromroomID, toroomID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error validating room link: " + err.Error())
		return
	}
	if linked {
		s.ChannelMessageSend(m.ChannelID, "Rooms already linked!")
		return
	}

	if direction == "up" {
		if toroom.DownID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room Down is already linked to: "+toroom.DownID)
			return
		}
		if fromroom.UpID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from Up is already linked to: "+fromroom.UpID)
			return
		}
		toroom.DownID = fromroomID
		fromroom.UpID = toroomID
	} else if direction == "down" {

		if toroom.UpID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room Up is already linked to: "+toroom.UpID)
			return
		}
		if fromroom.DownID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from Down is already linked to: "+fromroom.DownID)
			return
		}
		toroom.UpID = fromroomID
		fromroom.DownID = toroomID
	} else if direction == "north" {
		if toroom.SouthID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room South is already linked to: "+toroom.SouthID)
			return
		}
		if fromroom.NorthID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from North is already linked to: "+fromroom.NorthID)
			return
		}
		toroom.SouthID = fromroomID
		fromroom.NorthID = toroomID
	} else if direction == "northeast" {
		if toroom.SouthWestID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room SouthWest is already linked to: "+toroom.SouthWestID)
			return
		}
		if fromroom.NorthEastID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from NorthEast is already linked to: "+fromroom.NorthEastID)
			return
		}
		toroom.SouthWestID = fromroomID
		fromroom.NorthEastID = toroomID
	} else if direction == "east" {
		if toroom.WestID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room West is already linked to: "+toroom.WestID)
			return
		}
		if fromroom.EastID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from East is already linked to: "+fromroom.EastID)
			return
		}
		toroom.WestID = fromroomID
		fromroom.EastID = toroomID
	} else if direction == "southeast" {
		if toroom.NorthWestID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room NorthWest is already linked to: "+toroom.NorthWestID)
			return
		}
		if fromroom.SouthEastID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from SouthEast is already linked to: "+fromroom.SouthEastID)
			return
		}
		toroom.NorthWestID = fromroomID
		fromroom.SouthEastID = toroomID
	} else if direction == "south" {
		if toroom.NorthID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room North is already linked to: "+toroom.NorthID)
			return
		}
		if fromroom.SouthID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from South is already linked to: "+fromroom.SouthID)
			return
		}
		toroom.NorthID = fromroomID
		fromroom.SouthID = toroomID
	} else if direction == "southwest" {
		if toroom.NorthEastID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room NorthEast is already linked to: "+toroom.NorthEastID)
			return
		}
		if fromroom.SouthWestID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from SouthWest is already linked to: "+fromroom.SouthWestID)
			return
		}
		toroom.NorthEastID = fromroomID
		fromroom.SouthWestID = toroomID
	} else if direction == "west" {
		if toroom.EastID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room East is already linked to: "+toroom.EastID)
			return
		}
		if fromroom.WestID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from WestID is already linked to: "+fromroom.WestID)
			return
		}
		toroom.EastID = fromroomID
		fromroom.WestID = toroomID
	} else if direction == "northwest" {
		if toroom.SouthEastID != "" {
			s.ChannelMessageSend(m.ChannelID, "To room SouthEast is already linked to: "+toroom.SouthEastID)
			return
		}
		if fromroom.NorthEastID != "" {
			s.ChannelMessageSend(m.ChannelID, "From from NorthEast is already linked to: "+fromroom.NorthEastID)
			return
		}
		toroom.SouthEastID = fromroomID
		fromroom.NorthWestID = toroomID
	} else {
		s.ChannelMessageSend(m.ChannelID, "Unrecognized direction: "+direction)
		return
	}

	err = h.rooms.SaveRoomToDB(fromroom)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error updating DB: " + err.Error())
		return
	}

	err = h.rooms.SaveRoomToDB(toroom)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error updating DB: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Room " + fromroom.Name + " linked to " + toroom.Name)
	return
}

