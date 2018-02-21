package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"time"
)


type GuildsHandler struct {

	room 			*RoomsHandler
	guildmanager 	*GuildsManager
	registry		*CommandRegistry
	db				*DBHandler
	conf 			*Config
	perm			*PermissionsHandler
	user 			*UserHandler

	syncinprogress	bool

}


func (h *GuildsHandler) Init() {
	h.RegisterCommands()
	h.syncinprogress = false
}

func (h *GuildsHandler) RegisterCommands() (err error){

	h.registry.Register("guilds", "Manage rooms for this server", "sync | list | info | add | remove ")
	err = h.registry.AddGroup("guilds", "admin")
	return err

}

func (h *GuildsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.MainConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding usermanager")
		return
	}
	if !user.CheckRole("admin") {
		return
	}
	if strings.HasPrefix(m.Content, cp+"guilds") {
		if h.registry.CheckPermission("guilds", user, s, m) {

			// Grab our sender ID to verify if this usermanager has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving usermanager:" + m.Author.ID)
			}

			command := strings.Fields(m.Content)

			if user.CheckRole("admin") {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *GuildsHandler) ParseCommand(command []string, s *discordgo.Session, m *discordgo.MessageCreate) {


	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
		return
	}

	cmdlen := len(command)

	if cmdlen < 2 {
		s.ChannelMessageSend(m.ChannelID, "Expected flag for 'guilds' command, see usage for more info")
		return
	}
	if command[1] == "sync" {
		if cmdlen <= 2 {
			s.ChannelMessageSend(m.ChannelID, "sync requires at least one argument: roomroles | etc ")
			return
		}
		if cmdlen > 2 {
			if command[2] == "cluster"{
				// Syncs room roles across the entire cluster
				// This can take a long time so it's not recommended to run this which is why we look for a confirm response
				if cmdlen < 4 {
					s.ChannelMessageSend(m.ChannelID, "cluster-rooms requires you run the command with 'confirm' " +
						"before proceeding, note that this action may take a long time to complete depending on the size of your cluster")
					return
				}
				// Some people may be confused by the above so we'll take both inputs confirm and 'confirm' as valid
				if command[3] == "confirm" || command[3] == "'confirm'"{
					if h.syncinprogress {
						s.ChannelMessageSend(m.ChannelID, "Error syncing cluster: Sync already in progress, " +
							"please wait until the current one has completed!")
					}

					s.ChannelMessageSend(m.ChannelID, "Cluster sync started")
					h.syncinprogress = true
					starttime := time.Now()
					err := h.SyncCluster(s)
					if err != nil {
						h.syncinprogress = false
						s.ChannelMessageSend(m.ChannelID, "Error syncing cluster: " + err.Error())
					}
					delta := time.Since(starttime)
					h.syncinprogress = false
					s.ChannelMessageSend(m.ChannelID, "Cluster synced, took: " + delta.String())
					return
				} else {
					s.ChannelMessageSend(m.ChannelID, "invalid input, command canceled.")
					return
				}
			}
			if command[2] == "guild"{
				// Syncs room roles on the specified guild
				// This can take a long time so it's not recommended to run this which is why we look for a confirm response
				if cmdlen < 4 {
					s.ChannelMessageSend(m.ChannelID, "room-roles requires a guildID argument and that you run the command with 'confirm' " +
						"before proceeding, note that this action may take a long time to complete depending on the size of your cluster")
					return
				}
				if cmdlen < 5 {
					s.ChannelMessageSend(m.ChannelID, "guild requires you run the command with 'confirm' " +
						"before proceeding, note that this action may take a long time to complete depending on the size of your cluster")
					return
				}
				// Some people may be confused by the above so we'll take both inputs confirm and 'confirm' as valid
				if command[4] == "confirm" || command[4] == "'confirm'"{
					// Check if guildID is in the cluster list
					if h.guildmanager.IsGuildRegistered(command[3]){
						if h.syncinprogress {
							s.ChannelMessageSend(m.ChannelID, "Error syncing cluster: Sync already in progress, " +
								"please wait until the current one has completed!")
						}
						h.syncinprogress = true
						s.ChannelMessageSend(m.ChannelID, "Guild sync started")
						err := h.SyncGuild(command[3], s)
						if err != nil {
							h.syncinprogress = false
							s.ChannelMessageSend(m.ChannelID, "Error syncing guild: " + err.Error())
							return
						}
						h.syncinprogress = false
						s.ChannelMessageSend(m.ChannelID, "Guild synced: " + command[3])
						return
					} else {
						s.ChannelMessageSend(m.ChannelID, "Error: Supplied guild ID is not registered in the cluster.")
						return
					}
				} else {
					s.ChannelMessageSend(m.ChannelID, "invalid input, command canceled.")
					return
				}
			}
			s.ChannelMessageSend(m.ChannelID, "invalid argument specified, see command usage for help")
			return
		}
	}
	if command[1] == "info" {
		if cmdlen == 2 {
			// Display current guild info
			info, err := h.GuildInfo(guildID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not retrieve guild information: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, ":satellite: Guild Information: ```\n" + info + "\n```\n")
			return
		} else {
			if h.guildmanager.IsGuildRegistered(command[2]){
				// display guild info for supplied argument
				info, err := h.GuildInfo(command[2])
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not retrieve guild information: " + err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, ":satellite: Guild Information: ```\n" + info + "\n```\n")
				return
			} else {
				s.ChannelMessageSend(m.ChannelID, "Error: Supplied guild ID is not registered in the cluster.")
				return
			}
		}
	}
}




// This will sync the entire cluster, roles for every room and permissions for all of them
// This is a very intensive task so it's important that it not be run all the time
// There are timers throughout it to try and alleviate some of the strain on the api
func (h *GuildsHandler) SyncCluster(s *discordgo.Session) (err error){

	guilds, err := h.guildmanager.GetAllGuilds()
	if err != nil {
		return err
	}

	// Parse through guild list
	for _, guild := range guilds {
		// Wait 10 seconds between each guild
		time.Sleep(time.Duration(time.Second*10))
		h.SyncGuild(guild.ID, s)
	}

	return nil
}

// This will resync a specific guild overwriting any settings in the DB for it
// It will also fix roles for users in that guild
// It by itself will take a long time to finish
func (h *GuildsHandler) SyncGuild(guildID string, s *discordgo.Session) (err error){

	discordguild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	guild, err := h.guildmanager.GetGuildByID(guildID)
	if err != nil {
		return err
	}

	rooms, err := h.room.rooms.GetAllRooms()
	if err != nil {
		return err
	}

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	users, err := h.user.usermanager.GetAllUsers()
	if err != nil {
		return err
	}

	guild.Name = discordguild.Name
	guild.OwnerID = discordguild.OwnerID

	guild.AFKChannel = discordguild.AfkChannelID
	guild.AFKTimeout = discordguild.AfkTimeout
	guild.Icon = discordguild.Icon

	time.Sleep(time.Duration(time.Second*2))
	adminID, err := h.guildmanager.GetGuildDiscordAdminID(guild.ID, s)
	if err != nil {
		return err
	}
	guild.AdminID = adminID

	time.Sleep(time.Duration(time.Second*2))
	builderID, err := h.guildmanager.GetGuildDiscordBuilderID(guild.ID, s)
	if err != nil {
		return err
	}
	guild.BuilderID = builderID

	time.Sleep(time.Duration(time.Second*2))
	moderatorID, err := h.guildmanager.GetGuildDiscordModeratorID(guild.ID, s)
	if err != nil {
		return err
	}
	guild.ModeratorID = moderatorID

	time.Sleep(time.Duration(time.Second*2))
	everyoneID, err := h.guildmanager.GetGuildDiscordEveryoneID(guild.ID, s)
	if err != nil {
		return err
	}
	guild.EveryoneID = everyoneID

	err = h.guildmanager.SaveGuildToDB(guild)
	if err != nil {
		return err
	}

	// Parse through room list
	for _, room := range rooms {

		// If the room is in the current guild
		if room.GuildID == guild.ID {
			time.Sleep(time.Duration(time.Second*5))
			// Once per room we run syncroom which handles default permission assignments
			err = h.room.SyncRoom(room.ID, s)
			if err != nil {
				return err
			}
		}
	}

	for _, role := range roles {
		err = h.guildmanager.AddRoleToGuild(guildID, role.ID)
		if err != nil {
			return err
		}
	}

	for _, member := range discordguild.Members {
		err = h.guildmanager.AddUserToGuild(guildID, member.User.ID)
		if err != nil {
			return err
		}

		for _, user := range users {
			if user.GuildID == guildID {
				// Don't run a user repair for a user that is not in a channel
				if user.RoomID != "" {
					// Verify channel is in this guild
					for _, room := range discordguild.Channels {
						// If the user is in the room
						if room.ID == user.RoomID {
							time.Sleep(time.Duration(time.Second*1))
							err = h.user.RepairUser(user.ID, s, room.ID, user.GuildID)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}

	}
	return nil
}

func (h *GuildsHandler) GuildInfo(guildID string) (formatted string, err error){

	_, err = h.guildmanager.GetGuildByID(guildID)
	if err != nil {
		return "", err

	}
	return "", nil
}

