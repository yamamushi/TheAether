package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"errors"
	"strconv"
	"time"
	"sync"
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
	guilds 	 *GuildsManager

	roomsynclocker sync.RWMutex

}

func (h* RoomsHandler) InitRooms(s *discordgo.Session, channelID string) (err error){

	fmt.Println("Running Base Room Initialization")
	guildID := h.conf.MainConfig.CentralGuildID

	h.rooms = new(Rooms)
	h.rooms.db = h.db

	h.RegisterCommands()

	// Create default registered usermanager role
	registeredperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Registered", guildID, false, false, 16777215, registeredperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}

	// Spoilers Role
	spoilerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Spoilers", guildID, false, false, 16777215, spoilerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}


	fmt.Println("Creating Management Rooms")
	err = h.CreateManagementRooms(guildID, s)
	if err != nil {
		return err
	}

	fmt.Println("Creating OOC Rooms")
	err = h.CreateOOCChannels(guildID, s)
	if err != nil {
		return err
	}


	// This should now be handled by AddRoom automatically
	// Create default crossroads location role
	/*
	crossroadsperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Crossroads", guildID, true, false, 16747776, crossroadsperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	*/

/*
	theAetherRoleID, err := getRoleIDByName(s, guildID, "TheAether")
	theAetherPerms := h.perm.CreatePermissionInt(RolePermissions{CREATE_INSTANT_INVITE: true,
		KICK_MEMBERS: true, BAN_MEMBERS: true, ADMINISTRATOR: true, MANAGE_CHANNELS: true,
		MANAGE_GUILD: true, ADD_REACTIONS: true, VIEW_AUDIT_LOG: true, VIEW_CHANNEL: true,
		SEND_MESSAGES: true, SEND_TTS_MESSAGES: true, MANAGE_MESSAGES: true, EMBED_LINKS: true, ATTACH_FILES: true,
		READ_MESSAGE_HISTORY: true, MENTION_EVERYONE: true, USE_EXTERNAL_EMOJIS: true, CONNECT: true,
		SPEAK: true, MUTE_MEMBERS: true, DEAFEN_MEMBERS: true, MOVE_MEMBERS: true, USE_VAD: true,
		CHANGE_NICKNAME: true, MANAGE_NICKNAMES: true, MANAGE_ROLES: true, MANAGE_WEBHOOKS: true, MANAGE_EMOJIS: true})
	_, err = s.GuildRoleEdit(guildID, theAetherRoleID, "TheAether", 0, true, theAetherPerms, false)
	if err != nil {
		return err
	}
*/
	// The default Welcome Channel -> To be setup correctly a server NEEDS this channel and name as the default channel
	fmt.Println("Creating Lobby Rooms")
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

	registeredroleID, err := getRoleIDByName(s, guildID, "Registered")
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
	fmt.Println("Creating Crossroads Room")
	_, err = h.AddRoom(s, "crossroads", guildID, "The Aether", "", "", 16744704, false)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	crossroadsChannelID, err := getGuildChannelIDByName(s, guildID, "crossroads")
	if err != nil {
		fmt.Println("GetChannelIDByName: " + err.Error())
		return err
	}
	err = h.MoveRoom(s, crossroadsChannelID, guildID, "The Aether")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err
		}
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

	// This should now be handled by AddRoom
	/*
	crossroadsRoleID, err := getRoleIDByName(s, guildID, "Crossroads")
	if err != nil {
		return err
	}
	denyrossroadsperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowcrossroadsperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true})
	err = s.ChannelPermissionSet( crossroadsChannelID, crossroadsRoleID, "role", allowcrossroadsperms, denyrossroadsperms)
	if err != nil {
		return err
	}
	*/
/*
	room, err := h.rooms.GetRoomByID(crossroadsChannelID)
	if err != nil {
		room = Room{}
		room.ID = crossroadsChannelID
		room.RoleIDs = append(room.RoleIDs, crossroadsRoleID)
		room.TravelRoleID = crossroadsRoleID
		room.GuildID = guildID
		room.Name = "crossroads"
		err = h.rooms.SaveRoomToDB(room)
		if err != nil {
			fmt.Println("Crossroads SaveRoomToDB: " + err.Error())

			return err
		}
	} else {
		updatecrossroads := true
		for _, roleid := range room.RoleIDs {
			if roleid == crossroadsRoleID {
				updatecrossroads = false
			}
		}
		if updatecrossroads{
			room.RoleIDs = append(room.RoleIDs, crossroadsRoleID)
			room.TravelRoleID = crossroadsRoleID
			err = h.rooms.SaveRoomToDB(room)
			if err != nil {
				fmt.Println("SaveRoomToDB: " + err.Error())

				return err
			}
		}
	}
*/


	fmt.Println("Reordering Roles")
	err = h.perm.GuildReorderRoles(guildID, s)
	if err != nil {
		return err
	}

	return nil
}

// RegisterCommands function
func (h *RoomsHandler) RegisterCommands() (err error) {

	h.registry.Register("room", "Manage rooms for this server", "")
	h.registry.AddGroup("room", "builder")
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
		//fmt.Println("Error finding usermanager")
		return
	}
	if !user.CheckRole("builder") {
		return
	}
	if strings.HasPrefix(m.Content, cp+"room") {
		if h.registry.CheckPermission("room", user, s, m) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this usermanager has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving usermanager:" + m.Author.ID)
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
			s.ChannelMessageSend(m.ChannelID, "add requires at least one argument: <name> "+
				"<color (optional)>\nIf you would like to add a transfer room, the " +
					"syntax is: \n<name> <guildInviteLink> <transferRoomID> <color (optional)>" )
			return
		}

		parentname := "The Aether"
		transferID := ""
		transferRoomID := ""
		color := 0
		if len(command) == 4 {
			color, err = strconv.Atoi(command[3])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid color choice (must be integer)")
				return
			}
		}
		if len(command) == 5 {
			transferID = command[3]
			transferRoomID = command[4]
			return
		}
		if len(command) == 6 {
			transferID = command[3]
			transferRoomID = command[4]
			color, err = strconv.Atoi(command[5])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Invalid color choice (must be integer)")
				return
			}
		}
		if len(command) > 6 {
			s.ChannelMessageSend(m.ChannelID, "adding a transfer room requires three argument: <name> " +
				"<guildInviteLink> <transferRoomID> <color (optional)> ")
			return
		}

		channel, err := h.AddRoom(s, command[2], guildID, parentname, transferID, transferRoomID, color, false)
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

		roomname := command[2]
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

		s.ChannelMessageSend(m.ChannelID, "Channel " + roomname + " removed.")
		return
	}

	if command[1] == "linkrole" {
		if len(command) < 4 {
			s.ChannelMessageSend(m.ChannelID, "linkrole requires two arguments - <rolename> <#room>")
			return
		}

		h.LinkRole(command[2], command[3], s, m)
		return
	}
	if command[1] == "unlinkrole" {
		if len(command) < 4 {
			s.ChannelMessageSend(m.ChannelID, "unlinkrole requires two arguments - <rolename> <#room>")
			return
		}

		h.UnLinkRole(command[2], command[3], s, m)
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
			s.ChannelMessageSend(m.ChannelID, "Server configuration complete.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "setupserver requires an acknowledgement flag (y/n)")
			return
		}

	}

	if command[1] == "description" {
		if len(command) == 3 {
			description, err := h.GetRoomDescription(command[2])
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Error retrieving description: " + description)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room: "+ command[2] + " description: ```\n"+ description+"\n```\n")
			return
		}
		if len(command) >= 4 {

			description := ""
			for i, text := range command {
				if i > 2 {
					description = description + text + " "
				}
			}

			err := h.SetRoomDescription(command[2], description, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error setting description: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room description set: \n" + description)
			return
		}
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "description requires one or two arguments - <#room> <description>")
			return
		}

		h.LinkRole(command[2], command[3], s, m)
		return
	}

	if command[1] == "guildinvite" {
		if len(command) == 3 {
			invite, err := h.GetRoomTransferInvite(command[2])
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Error retrieving guild invite: " + invite)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room: "+ command[2] + " invite: ```\n"+ invite+"\n```\n")
			return
		}
		if len(command) >= 4 {

			err := h.SetRoomTransferInvite(command[2], command[3], s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error setting description: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room invite set: \n" + command[3])
			return
		}
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "guildinvite requires one or two arguments - <#room> <invite>")
			return
		}

		h.LinkRole(command[2], command[3], s, m)
		return
	}

	// list roles for a room
	if command[1] == "roles" {
		if len(command) < 3 {
			s.ChannelMessageSend(m.ChannelID, "roles requires an argument - <#room>")
			return
		}

		roomID := CleanChannel(command[2])

		roles, err := h.GetRoomRoles(roomID, guildID, s)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve roles for room: " + err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Roles for " + command[2] + " : " + roles )
		return
	}

	// Set and unset travel role
	if command[1] == "travelrole" {
		if len(command) > 3 {
			err := h.SetRoomTravelRole(command[3], command[2], guildID, s)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error setting description: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room travel role set.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "travelroleclear requires two arguments - <#room> <rolename>")
			return
		}
		return
	}
	if command[1] == "travelroleclear" {
		if len(command) > 2 {
			err := h.RemoveRoomTravelRole(command[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error clearing travel role: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room travel role cleared.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "travelroleclear requires an argument - <#room>")
			return
		}
		return
	}

	// Set and unset transfer role
	if command[1] == "transferrole" {
		if len(command) > 3 {
			err := h.SetRoomTransferRoleID(command[2], command[3], s)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error setting transfer role: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room transfer role set.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "transferrole requires two arguments - <#room> <roleID>")
			return
		}
		return
	}
	if command[1] == "transferroleclear" {
		if len(command) > 2 {
			err := h.RemoveRoomTransferRoleID(command[2])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error clearing transfer role: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Room transfer role cleared.")
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "transferroleclear requires an argument - <#room>")
			return
		}
		return
	}

}



// CreateManagementRooms function - Useful for creating default management roles and rooms for new guilds
func (h *RoomsHandler) CreateManagementRooms(guildID string, s *discordgo.Session) (err error){

	// Create default developers role
	developerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Developer", guildID, true, false, 255, developerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default admin role
	adminsperms := h.perm.CreatePermissionInt(RolePermissions{ADMINISTRATOR:true})
	_, err = h.perm.CreateRole("Admin", guildID, true, true, 16767744, adminsperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default builder role
	builderperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true})
	_, err = h.perm.CreateRole("Builder", guildID, true, false, 11993343, builderperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default moderator role
	moderatorperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, SEND_MESSAGES: true,
		READ_MESSAGE_HISTORY: true, MANAGE_MESSAGES: true, KICK_MEMBERS: true, BAN_MEMBERS: true})
	_, err = h.perm.CreateRole("Moderator", guildID, true, false, 38659, moderatorperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Create default writer role
	writerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Writer", guildID, true, false, 16750591, writerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	// Developers
	_, err = h.AddRoom(s, "developers", guildID, "Management", "", "", 0, true)
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
	allowdevperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, READ_MESSAGE_HISTORY:true, SEND_MESSAGES:true,
										USE_EXTERNAL_EMOJIS:true, ATTACH_FILES:true, EMBED_LINKS:true, MENTION_EVERYONE:true})
	err = s.ChannelPermissionSet( developerChannelID, developerRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}


	// Admin
	_, err = h.AddRoom(s, "admins", guildID, "Management", "", "", 0, true)
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
	allowadminperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true,
		EMBED_LINKS:true, READ_MESSAGE_HISTORY:true, ATTACH_FILES:true, USE_EXTERNAL_EMOJIS:true, MENTION_EVERYONE:true})
	err = s.ChannelPermissionSet( adminChannelID, adminRoleID, "role", allowadminperms, denyadminperms)
	if err != nil {
		return err
	}


	// Builder
	_, err = h.AddRoom(s, "builders", guildID, "Management", "", "", 0, true)
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
	allowbuilderperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true,
						EMBED_LINKS:true, READ_MESSAGE_HISTORY:true, ATTACH_FILES:true, USE_EXTERNAL_EMOJIS:true, MENTION_EVERYONE:true})
	err = s.ChannelPermissionSet( builderhannelID, builderRoleID, "role", allowbuilderperms, denybuilderperms)
	if err != nil {
		return err
	}


	// Moderator
	_, err = h.AddRoom(s, "moderators", guildID, "Management", "", "", 0, true)
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
	allowmoderatorperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true,
		EMBED_LINKS:true, READ_MESSAGE_HISTORY:true, ATTACH_FILES:true, USE_EXTERNAL_EMOJIS:true, MENTION_EVERYONE:true})
	err = s.ChannelPermissionSet( moderatorhannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}


	// Writer
	_, err = h.AddRoom(s, "writers", guildID, "Management", "", "", 0, true )
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
	allowwriterperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true,
		EMBED_LINKS:true, READ_MESSAGE_HISTORY:true, ATTACH_FILES:true, USE_EXTERNAL_EMOJIS:true, MENTION_EVERYONE:true})
	err = s.ChannelPermissionSet( writerchannelID, writerRoleID, "role", allowwriterperms, denywriterperms)
	if err != nil {
		return err
	}

	return nil
}

func (h *RoomsHandler) CreateDefaultRoles(guildID string, s *discordgo.Session) (err error){
	// Create default registered usermanager role
	registeredperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Registered", guildID, false, false, 16777215, registeredperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}

	spoilerperms := h.perm.CreatePermissionInt(RolePermissions{})
	_, err = h.perm.CreateRole("Spoilers", guildID, false, false, 16777215, spoilerperms, s)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}


	// The default Welcome Channel -> To be setup correctly a server NEEDS this channel and name as the default channel
	fmt.Println("Creating Lobby Rooms")
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
	denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{SEND_MESSAGES: true})
	alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{READ_MESSAGE_HISTORY:true, VIEW_CHANNEL: true})
	err = s.ChannelPermissionSet( welcomeChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}

	registeredroleID, err := getRoleIDByName(s, guildID, "Registered")
	if err != nil {
		return err
	}
	denyregisteredperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	allowregisteredperms := h.perm.CreatePermissionInt(RolePermissions{})
	err = s.ChannelPermissionSet( welcomeChannelID, registeredroleID, "role", allowregisteredperms, denyregisteredperms)
	if err != nil {
		return err
	}

	return nil
}

func (h *RoomsHandler) CreateOOCChannels(guildID string, s *discordgo.Session) (err error){

	everyoneID, err := getGuildEveryoneRoleID(s, guildID)
	if err != nil {
		return err
	}

	denyeveryoneperms := h.perm.CreatePermissionInt(RolePermissions{SEND_MESSAGES: true})
	alloweveryoneperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, READ_MESSAGE_HISTORY: true})


	// rules
	_, err = h.AddRoom(s, "rules", guildID, "Lobby", "", "", 0, true)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	rulesChannelID, err := getGuildChannelIDByName(s, guildID, "rules")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, rulesChannelID, guildID, "Lobby")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( rulesChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	ruleschannelEdit := new(discordgo.ChannelEdit)
	ruleschannelEdit.Topic = "Rules - The Aether v1.0"
	ruleschannelEdit.Position = 1
	_, err = s.ChannelEditComplex(rulesChannelID, ruleschannelEdit)
	if err != nil {
		return err
	}


	denyeveryoneperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true})
	alloweveryoneperms = h.perm.CreatePermissionInt(RolePermissions{})

	registeredRoleID, err := getRoleIDByName(s, guildID, "Registered")
	if err != nil {
		return err
	}
	denydevperms := h.perm.CreatePermissionInt(RolePermissions{})
	allowdevperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true, READ_MESSAGE_HISTORY: true, USE_EXTERNAL_EMOJIS:true})


	moderatorRoleID, err := getRoleIDByName(s, guildID, "Moderator")
	if err != nil {
		return err
	}
	allowmoderatorperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, SEND_MESSAGES: true,
		READ_MESSAGE_HISTORY: true, MANAGE_MESSAGES: true, KICK_MEMBERS: true, BAN_MEMBERS: true})
	denymoderatorperms := h.perm.CreatePermissionInt(RolePermissions{})

	// ooc
	_, err = h.AddRoom(s, "ooc", guildID, "OOC", "", "", 0, true)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	oocChannelID, err := getGuildChannelIDByName(s, guildID, "ooc")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, oocChannelID, guildID, "OOC")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( oocChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( oocChannelID, registeredRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( oocChannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}
	oocChannelEdit := new(discordgo.ChannelEdit)
	oocChannelEdit.Topic = "Out of Character Chat - DO NOT Discuss spoilers here!"
	oocChannelEdit.Position = 0
	_, err = s.ChannelEditComplex(oocChannelID, oocChannelEdit)
	if err != nil {
		return err
	}


	// trades
	_, err = h.AddRoom(s, "trades", guildID, "OOC", "", "", 0, true )
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	tradesChannelID, err := getGuildChannelIDByName(s, guildID, "trades")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, tradesChannelID, guildID, "OOC")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( tradesChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( tradesChannelID, registeredRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( tradesChannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}
	tradeschannelEdit := new(discordgo.ChannelEdit)
	tradeschannelEdit.Topic = "Trades Chat - Please only post buy and sell orders here, take discussions to private chat!"
	tradeschannelEdit.Position = 2
	_, err = s.ChannelEditComplex(tradesChannelID, tradeschannelEdit)
	if err != nil {
		return err
	}



	// help
	_, err = h.AddRoom(s, "help", guildID, "OOC", "", "", 0, true)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	helpChannelID, err := getGuildChannelIDByName(s, guildID, "help")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, helpChannelID, guildID, "OOC")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( helpChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( helpChannelID, registeredRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( helpChannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}
	helpchannelEdit := new(discordgo.ChannelEdit)
	helpchannelEdit.Topic = "Help Chat - No Spoilers! Get help on using game commands here"
	helpchannelEdit.Position = 3
	_, err = s.ChannelEditComplex(helpChannelID, helpchannelEdit)
	if err != nil {
		return err
	}


	// spoilers
	// We want this role to be created
	_, err = h.AddRoom(s, "spoilers", guildID, "OOC", "", "", 0, false )
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	spoilersChannelID, err := getGuildChannelIDByName(s, guildID, "spoilers")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, tradesChannelID, guildID, "OOC")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( spoilersChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( spoilersChannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}
	spoilerschannelEdit := new(discordgo.ChannelEdit)
	spoilerschannelEdit.Topic = "Spoilers Chat - You have been warned! Do not post spoilers outside of this channel!"
	spoilerschannelEdit.Position = 4
	_, err = s.ChannelEditComplex(spoilersChannelID, spoilerschannelEdit)
	if err != nil {
		return err
	}


	// bugs
	_, err = h.AddRoom(s, "bugs", guildID, "OOC", "", "", 0, true)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists"){
			return err
		}
	}
	bugsChannelID, err := getGuildChannelIDByName(s, guildID, "bugs")
	if err != nil {
		return err
	}
	err = h.MoveRoom(s, bugsChannelID, guildID, "OOC")
	if err != nil {
		if !strings.Contains(err.Error(), "No record found"){
			return err // We don't care about no record being found in the Lobby because it is our default room
		}
	}
	err = s.ChannelPermissionSet( bugsChannelID, everyoneID, "role", alloweveryoneperms, denyeveryoneperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( bugsChannelID, registeredRoleID, "role", allowdevperms, denydevperms)
	if err != nil {
		return err
	}
	err = s.ChannelPermissionSet( bugsChannelID, moderatorRoleID, "role", allowmoderatorperms, denymoderatorperms)
	if err != nil {
		return err
	}
	bugschannelEdit := new(discordgo.ChannelEdit)
	bugschannelEdit.Topic = "Bugs Chat - Please provide as much details as you can, or post an issue on github!"
	bugschannelEdit.Position = 5
	_, err = s.ChannelEditComplex(bugsChannelID, bugschannelEdit)
	if err != nil {
		return err
	}

	return nil
}

func (h *RoomsHandler) DeleteRoom(s *discordgo.Session, name string, guildID string, parentname string) {



}

func (h *RoomsHandler) AddRoom(s *discordgo.Session, name string, guildID string, parentname string,
	transferInvite string, transferRoomID string, color int, overriderole bool) (createdroom *discordgo.Channel, err error) {

	rooms, err := h.rooms.GetAllRooms()
	if err != nil {
		return createdroom, err
	}
	if len(rooms) >= 80 {
		return createdroom, errors.New("Maximum supported rooms reached!")
	}

	if transferRoomID != "" {
		_, err := s.Channel(transferRoomID)
		if err != nil {
			return createdroom, errors.New("Could not find target transfer room: " + err.Error())
		}
	}

	record, err := h.rooms.GetRoomByName(name, guildID)
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

		record = Room{ID: createdchannel.ID, GuildID: guildID, Name: name, ParentID: parentID, ParentName: parentname,
		GuildTransferInvite: transferInvite, TransferRoomID: transferRoomID}

		// Everyone Default Roles
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


		builderID, err := getRoleIDByName(s, guildID, "Builder")
		if err != nil {
			return createdroom, err
		}
		denybuilder := h.perm.CreatePermissionInt(RolePermissions{})
		allowbuilder := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, SEND_MESSAGES: true})
		err = s.ChannelPermissionSet( createdroom.ID, builderID, "role", allowbuilder, denybuilder)
		if err != nil {
			return createdroom, err
		}


		moderatorID, err := getRoleIDByName(s, guildID, "Moderator")
		if err != nil {
			return createdroom, err
		}
		denymoderator := h.perm.CreatePermissionInt(RolePermissions{})
		allowmoderator := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL: true, SEND_MESSAGES: true,
		READ_MESSAGE_HISTORY: true, MANAGE_MESSAGES: true, KICK_MEMBERS: true, BAN_MEMBERS: true})
		err = s.ChannelPermissionSet( createdroom.ID, moderatorID, "role", allowmoderator, denymoderator)
		if err != nil {
			return createdroom, err
		}


		// Create default role here
		if !overriderole {
			createdrole, err := h.perm.CreateRole(name, guildID, true, false, color, 0, s)
			newroleID := ""
			if err != nil {
				if !strings.Contains(err.Error(), "already exists"){
					return createdroom, err
				} else {
					newroleID, err 	= getRoleIDByName(s, guildID, name)
					if err != nil {
						return createdroom, err
					}
				}
			} else {
				newroleID = createdrole.ID
			}

			record.AdditionalRoleIDs = append(record.AdditionalRoleIDs, newroleID)
			record.TravelRoleID = newroleID

			denyrperms := 0
			allowperms := 0
			if record.GuildTransferInvite != "" {
				denyrperms = h.perm.CreatePermissionInt(RolePermissions{SEND_MESSAGES:true, READ_MESSAGE_HISTORY:true})
				allowperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, READ_MESSAGE_HISTORY:true, USE_EXTERNAL_EMOJIS:true})
			} else {
				denyrperms = h.perm.CreatePermissionInt(RolePermissions{})
				allowperms = h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true})
			}
			err = s.ChannelPermissionSet( record.ID, newroleID, "role", allowperms, denyrperms)
			if err != nil {
				return createdroom, err
			}

			err = h.perm.GuildReorderRoles(guildID, s)
			if err != nil {
				return createdroom, err
			}
		}

		h.rooms.SaveRoomToDB(record)
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


	if existingrecord.DownID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.DownID)
		if err == nil {
			linkedroom.UpID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.NorthID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.NorthID)
		if err == nil {
			linkedroom.SouthID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.NorthEastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.NorthEastID)
		if err == nil {
			linkedroom.SouthWestID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.EastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.EastID)
		if err == nil {
			linkedroom.WestID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.SouthEastID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.SouthEastID)
		if err == nil {
			linkedroom.NorthWestID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.SouthID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.SouthID)
		if err == nil {
			linkedroom.NorthID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.SouthWestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.SouthWestID)
		if err == nil {
			linkedroom.NorthEastID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.WestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.WestID)
		if err == nil {
			linkedroom.EastID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	if existingrecord.NorthWestID != "" {
		linkedroom, err := h.rooms.GetRoomByID(existingrecord.NorthWestID)
		if err == nil {
			linkedroom.SouthEastID = ""
			err = h.rooms.SaveRoomToDB(linkedroom)
			if err != nil {
				return err
			}
		}
	}

	// Now clean up travel records
	rooms, err := h.rooms.GetAllRooms()
	if err != nil{
		return err
	}

	for _, search := range rooms {
		if search.TransferRoomID != "" {
			// If the search transfer room ID points to this room
			// update the search record so that it is not anymore
			if search.TransferRoomID == existingrecord.ID {
				search.TransferRoomID = ""
				search.GuildTransferInvite = ""

				err = h.rooms.SaveRoomToDB(search)
				if err != nil {
					return err
				}
				err = h.perm.ApplyTravelRolePerms(search.ID, search.GuildID, s)
				if err != nil {
					return err
				}
			}
		}
	}


	s.ChannelDelete(existingrecord.ID)

	err = h.registry.RemoveChannel("travel", existingrecord.ID)
	if err != nil {
		return err
	}


	err = h.rooms.RemoveRoomFromDB(existingrecord)
	if err != nil {
		return err
	}

	return nil
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

func (h *RoomsHandler) AddUserIDToRoomRecord(userID string, roomID string, guildID string, s *discordgo.Session)(err error){

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	_, err = h.user.GetUser(userID, s, guildID)
	if err != nil {
		return err
	}

	for _, roomUserID := range room.UserIDs {
		if userID == roomUserID {
			fmt.Println("User already in room!: " + userID + " size of rooms: " + strconv.Itoa(len(room.UserIDs)))
			return nil // already in record
		}
	}

	room.UserIDs = append(room.UserIDs, userID)

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		return err
	}
	//fmt.Println("Added userid to record: " + userID + " size of rooms: " + strconv.Itoa(len(room.UserIDs)))
	return nil
}

func (h *RoomsHandler) RemoveUserIDFromRoomRecord(userID string, roomID string)(err error){

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	room.UserIDs = RemoveStringFromSlice(room.UserIDs, userID)

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) SetupNewServer(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	if guildID == h.conf.MainConfig.CentralGuildID {
		return errors.New("This cannot be run on the central guild!")
	}

	err = h.CreateDefaultRoles( guildID, s )
	if err != nil {
		return err
	}
	err = h.CreateManagementRooms(guildID, s)
	if err != nil {
		return err
	}
	/*
	err = h.CreateOOCChannels(guildID, s)
	if err != nil {
		return err
	}
	*/
	err = h.guilds.RegisterGuild(guildID, s)
	if err != nil {
		return err
	}

	err = h.perm.GuildReorderRoles(guildID, s)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) SetTravelRole(rolename string, roomID string, s *discordgo.Session,  m *discordgo.MessageCreate) (err error) {

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return errors.New("Error retrieving roleID: " + err.Error())
	}

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		return errors.New("Could not retrieve GuildID: " + err.Error())
	}

	roleID, err := getRoleIDByName(s, guildID, rolename)
	if err != nil {
		return errors.New("Error retrieving roleID: " + err.Error())
	}

	_, err = s.Channel(roomID)
	if err != nil {
		return errors.New("Could not find target transfer room: " + err.Error())
	}

	room.TravelRoleID = roleID

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		return errors.New("Error updating DB: " + err.Error())
	}

	denyrperms := 0
	allowperms := 0
	if room.GuildTransferInvite != "" {
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

	return nil
}

func (h *RoomsHandler) SyncRolePermissions(roomID string) (err error) {

	return nil
}


// Linking Roles

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

	for _, additionalID := range room.AdditionalRoleIDs {
		if additionalID == roleID {
			s.ChannelMessageSend(m.ChannelID, "Room is already linked to role!")
			return
		}
	}
	room.AdditionalRoleIDs = append(room.AdditionalRoleIDs, roleID)

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error updating DB: " + err.Error())
		return
	}

	denyrperms := 0
	allowperms := 0
	if room.GuildTransferInvite != "" {
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

func (h *RoomsHandler) UnLinkRole(rolename string, roomID string, s *discordgo.Session, m *discordgo.MessageCreate) {

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

	room.AdditionalRoleIDs = RemoveStringFromSlice(room.AdditionalRoleIDs, roleID)

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error updating DB: " + err.Error())
		return
	}

	denyperms := h.perm.CreatePermissionInt(RolePermissions{VIEW_CHANNEL:true, SEND_MESSAGES: true})

	err = s.ChannelPermissionSet( room.ID, roleID, "role", 0, denyperms)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error setting permissions: " + err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Role " + rolename + " linked to " + room.Name)
	return
}


// Room Info
func (h *RoomsHandler) FormatRoomInfo(roomID string) (formatted string, err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return "", err
	}

	output := "\n```\n"
	output = output + "Name: " + room.Name + "\n"
	output = output + "ID: " +  room.ID + "\n"
	output = output + "Guild: " + room.GuildID + "\n\n"
	output = output + "GuildTransferInvite: " + room.GuildTransferInvite + "\n"
	output = output + "TransferRoomID: " + room.TransferRoomID + "\n\n"
	output = output + "ParentID: " + room.ParentID + "\n"
	output = output + "ParentName: " + room.ParentName + "\n\n"
	output = output + "TravelRoleID: " + room.TravelRoleID + "\n"
	output = output + "Current User Count: " + strconv.Itoa(len(room.UserIDs)) + "\n"
	roles := ""
	for i, additionalrole := range room.AdditionalRoleIDs {
		if i > 0 {
			roles = ", " + additionalrole
		} else {
			roles = additionalrole
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


// Directional Commands
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


// Get Set for various room details

func (h *RoomsHandler) GetRoomDescription(roomID string) (formatted string, err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return formatted, nil
	}

	formatted = room.Description
	return formatted, nil
}

func (h *RoomsHandler) SetRoomDescription(roomID string, description string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	room.Description = description

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) GetRoomTransferInvite(roomID string) (formatted string, err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return formatted, nil
	}

	formatted = room.GuildTransferInvite
	return formatted, nil
}

func (h *RoomsHandler) SetRoomTransferInvite(roomID string, invite string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	roomID = CleanChannel(roomID)

	if !strings.HasPrefix(invite, "http"){
		return errors.New("Invite url is not formatted correctly")
	}

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	room.GuildTransferInvite = invite

	err = h.rooms.SaveRoomToDB(room)
	if err != nil {
		return err
	}

	return nil
}


func (h *RoomsHandler) GetRoomRoles(roomID string, guildID string, s *discordgo.Session) (formatted string, err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return "", err
	}

	formatted = "```\nTravelRoleID: "+room.TravelRoleID+"\n\nRole List: \n"

	for _, roleID := range room.AdditionalRoleIDs {

		rolename, err := getRoleNameByID(roleID, guildID, s)
		if err != nil {
			return "", err
		}
		formatted = formatted + rolename + ": " + roleID + "\n"
	}

	formatted = formatted + "```\n"
	return formatted, nil
}


func (h *RoomsHandler) SetRoomTravelRole(rolename string, roomID string, guildID string, s *discordgo.Session) (err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	roleID, err := getRoleIDByName(s, guildID, rolename)
	if err != nil {
		return err
	}

	room.TravelRoleID = roleID

	return h.rooms.SaveRoomToDB(room)
}

func (h *RoomsHandler) RemoveRoomTravelRole(roomID string) (err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	room.TravelRoleID = ""

	return h.rooms.SaveRoomToDB(room)

}


func (h *RoomsHandler) SetRoomTransferRoleID(roleID string, roomID string, s *discordgo.Session) (err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	_, err = s.Channel(roleID)
	if err != nil {
		return errors.New("Could not find target transfer room: " + err.Error())
	}


	room.TravelRoleID = roleID

	return h.rooms.SaveRoomToDB(room)
}

func (h *RoomsHandler) RemoveRoomTransferRoleID(roomID string) (err error) {

	roomID = CleanChannel(roomID)

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	room.TransferRoomID = ""

	return h.rooms.SaveRoomToDB(room)

}




func (h *RoomsHandler) SyncRoom(roomID string, s *discordgo.Session) (err error){
	h.roomsynclocker.Lock() // One room at a time!
	defer h.roomsynclocker.Unlock()

	room, err := h.rooms.GetRoomByID(roomID)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(time.Second*2))
	adminID, err := h.guilds.GetGuildDiscordAdminID(room.GuildID, s)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(time.Second*2))
	builderID, err := h.guilds.GetGuildDiscordBuilderID(room.GuildID, s)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(time.Second*2))
	moderatorID, err := h.guilds.GetGuildDiscordModeratorID(room.GuildID, s)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(time.Second*2))
	everyoneID, err := h.guilds.GetGuildDiscordEveryoneID(room.GuildID, s)
	if err != nil {
		return err
	}

	err = h.perm.ApplyAdminRolePerms(room.ID, room.GuildID, adminID, s)
	if err != nil {
		return err
	}

	err = h.perm.ApplyModeratorRolePerms(room.ID, room.GuildID, moderatorID, s)
	if err != nil {
		return err
	}

	err = h.perm.ApplyBuilderRolePerms(room.ID, room.GuildID, builderID, s)
	if err != nil {
		return err
	}

	err = h.perm.ApplyEveryoneRolePerms(room.ID, room.GuildID, everyoneID, s)
	if err != nil {
		return err
	}

	// For every room we go through its role id list
	for _, roleID := range room.AdditionalRoleIDs {

		// Wait 3 seconds for each role in the room
		time.Sleep(time.Duration(time.Second*3))
		if roleID == room.TravelRoleID {
			err = h.perm.ApplyTravelRolePerms(room.ID, room.GuildID, s)
			if err != nil {
				return err
			}
		} else {
			// We don't have extra tasks for extra roles YET
		}
	}
	return nil
}
