package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"strconv"
	"time"
)

// PermissionsHandler struct
type PermissionsHandler struct {
	db       *DBHandler
	conf     *Config
	dg       *discordgo.Session
	callback *CallbackHandler
	user     *UserHandler
	logchan  chan string
	room     *RoomsHandler
}

// Read function
func (h *PermissionsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	// Verify the user account exists (creates one if it doesn't exist already)
	h.user.CheckUser(m.Author.ID, s, m.ChannelID )

	/*
		user, err := h.db.GetUser(m.Author.ID)
		if err != nil{
			fmt.Println("Error finding user")
			return
		}
	*/

	// Command prefix
	cp := h.conf.MainConfig.CP

	// Command from message content
	command := strings.Fields(m.Content)
	// We use this a bit, this is the author id formatted as a mention
	//authormention := m.Author.Mention()
	//mentions := m.Mentions

	// We don't care about commands that aren't formatted for this handler
	if len(command) < 1 {
		return
	}

	command[0] = strings.TrimPrefix(command[0], cp)

	// After our command string has been trimmed down, check it against the command list
	if command[0] == "perms" {
		if len(command) < 2 {
			s.ChannelMessageSend(m.ChannelID, "<perms> expects an argument.")
			return
		}

		guildID, err := getGuildID(s, m.ChannelID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not retrieve GuildID: " + err.Error())
			return
		}

		if command[1] == "addrole"{
			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) != 4 {
				s.ChannelMessageSend(m.ChannelID, "<addrole> expects two argument - <role name> <user mention>.")
				return
			}

			if len(m.Mentions) < 1 {
				s.ChannelMessageSend(m.ChannelID, "<addrole> expects two argument - <role name> <user mention>.")
				return
			}

			testrole := strings.ToLower(command[2])
			if testrole == "admin" {
				if !user.CheckRole("admin") {
					s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
					return
				}
			}

			err = h.AddRoleToUser( command[2], m.Mentions[0].ID, s, m)
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Error adding role: " + err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Role " + command[2] + " added to user")
			return
		}
		if command[1] == "removerole"{
			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) < 4{
				s.ChannelMessageSend(m.ChannelID, "<removerole> expects two argument - <role name> <user mention>.")
				return
			}

			if len(m.Mentions) < 1 {
				s.ChannelMessageSend(m.ChannelID, "<addrole> expects two argument - <role name> <user mention>.")
				return
			}

			testrole := strings.ToLower(command[2])
			if testrole == "admin" {
				if !user.CheckRole("admin") {
					s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
					return
				}
			}

			err = h.RemoveRoleFromUser( command[2], m.Mentions[0].ID, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error removing role: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Role " + command[2] + " removed from user")
			return
		}

		if command[1] == "createrole"{

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("admin") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) < 6 {
				s.ChannelMessageSend(m.ChannelID, "<createrole> expects 4 arguments - <role name> <hoist> <mentionable> <color int>")
				return
			}

			hoist := false
			if command[3] == "true" {
				hoist = true
			} else if command[3] == "false" {
				hoist = false
			} else {
				s.ChannelMessageSend(m.ChannelID, "<hoist> must be true or false)")
				return
			}

			mentionable := false
			if command[4] == "true" {
				mentionable = true
			} else if command[4] == "false" {
				mentionable = false
			} else {
				s.ChannelMessageSend(m.ChannelID, "<mentionable> must be true or false)")
				return
			}

			color := 0
			color, err = strconv.Atoi(command[5])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "<color> must be an integer)")
				return
			}

			perms := h.CreatePermissionInt(RolePermissions{})
			createdrole, err := h.CreateRole( command[2], guildID, hoist, mentionable, color, perms, s)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error creating role: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Role " + createdrole.Name + " created: " + createdrole.ID)
			return
		}
		if command[1] == "deleterole" {
			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("admin") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) < 3 {
				s.ChannelMessageSend(m.ChannelID, "<deleterole> expects an argument - <role name>")
				return
			}

			rolename := command[2]
			roleID, err := getRoleIDByName(s, guildID, rolename)
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Could not delete role: " + err.Error())
				return
			}

			err = h.DeleteRoleOnGuild(roleID, guildID, s)
			if err != nil{
				s.ChannelMessageSend(m.ChannelID, "Could not delete role: " + err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, "Role deleted on server!")
			return
		}

		if command[1] == "viewrole" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) < 3 {
				s.ChannelMessageSend(m.ChannelID, "<create> expects an argument - <role name> or <role id>")
				return
			}

			h.ViewRole(command[2], guildID, s, m)
			return
		}

		if command[1] == "promote" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			// Run our promote command function
			command = RemoveStringFromSlice(command, command[0])
			h.ReadPromote(command, s, m)
			return
		}
		if command[1] == "demote" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			// Run our promote command function
			command = RemoveStringFromSlice(command, command[0])
			h.ReadDemote(command, s, m)
			return
		}

		if command[1] == "syncserverroles" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("admin") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(m.Mentions) < 1 {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles expects a user mention!")
				return
			} else {
				err = h.SyncServerRoles(m.Mentions[0].ID, m.ChannelID, s)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
					return
				}
			}
			return
		}
		if command[1] == "syncrolesdb" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
				return
			}

			if !user.CheckRole("admin") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(m.Mentions) < 1 {
				s.ChannelMessageSend(m.ChannelID, "syncrolesdb expects a user mention!")
				return
			} else {
				err = h.SyncRolesDB(m.Mentions[0].ID, guildID, m.ChannelID, s)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
					return
				}
			}
			return
		}

		if command[1] == "translaterole" {

			// Get the authors user object from the database
			user, err := h.db.GetUser(m.Author.ID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "translaterole error: " + err.Error())
				return
			}

			if !user.CheckRole("moderator") {
				s.ChannelMessageSend(m.ChannelID, "You do not have permission to use this command.")
				return
			}

			if len(command) < 3 {
				s.ChannelMessageSend(m.ChannelID, "translaterole expects an argument: roleID")
				return
			} else {
				rolename, err := h.TranslateRoleID(command[2], guildID, s)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "syncserverroles error: " + err.Error())
					return
				}
				output := "RoleID Translation: ```\nRoleID: "+ command[2] + "\nRole Name: " + rolename + "\n```\n"

				s.ChannelMessageSend(m.ChannelID, output )
				return
			}
			return
		}
	}

	return
}

func (h *PermissionsHandler) ViewRole(rolename string, guildID string, s *discordgo.Session, m *discordgo.MessageCreate) {

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not retrieve guild roles: " + err.Error())
		return
	}

	formatted := "```\n"
	for _, role := range roles {
		if role.Name == rolename || role.ID == rolename {
			formatted = formatted + "Name: " + role.Name + "\n"
			formatted = formatted + "ID: " + role.ID + "\n"
			formatted = formatted + "Perms: " + strconv.Itoa(role.Permissions) + "\n"
			formatted = formatted + "Hoist: " + strconv.FormatBool(role.Hoist) + "\n"
			formatted = formatted + "Mentionable: " + strconv.FormatBool(role.Mentionable) + "\n"
			formatted = formatted + "Color: " + strconv.Itoa(role.Color) + "\n"
		}
	}
	formatted = formatted + "```\n"


	s.ChannelMessageSend(m.ChannelID, "Role: \n" + formatted)
	return
}




// ReadPromote The promote command runs using our commands array to get the promotion settings
func (h *PermissionsHandler) ReadPromote(commands []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(commands) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: promote <user> <group>")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "User must be mentioned")
		return
	}

	// Grab our target user id and group
	target := m.Mentions[0].ID
	group := commands[2]

	// Get the authors user object from the database
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println("Could not find user in PermissionsHandler.ReadPromote")
		return
	}

	// Check the group argument
	if group == "owner" {
		if !user.CheckRole("owner") {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" https://www.youtube.com/watch?v=fmz-K2hLwSI ")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to owner"
			return
		}
		s.ChannelMessageSend(m.ChannelID, "This group cannot be assigned through the promote command.")
		h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to owner"
		return

	}
	if group == "admin" {
		if !user.CheckRole("owner") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to admin"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to admin || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "smoderator" {

		if !user.CheckRole("admin") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to smoderator"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to smoderator || " +
				m.Mentions[0].Mention() + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "moderator" {

		if !user.CheckRole("smoderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to moderator"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to moderator || " +
				target + "||" + group + "||" + err.Error()

			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "builder" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to editor"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to editor || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "writer" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to agora"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to agora || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "scripter" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to streamer"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to streamer || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "architect" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to recruiter"
			return
		}
		err = h.Promote(target, group, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to recruiter || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been added to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been added to the " + group + " group by " + m.Author.Mention()
		return

	}
	s.ChannelMessageSend(m.ChannelID, group+" is not a valid group!")
	h.logchan <- "Permissions " + m.Author.Mention() + " attempted to promote " + m.Mentions[0].Mention() +
		" to " + group + " which does not exist"
	return

}

// Promote Set the given role on a user, and save the changes in the database
func (h *PermissionsHandler) Promote(userid string, group string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	// Get user from the database using the userid
	user, err := h.user.GetUser(userid, s, m.ChannelID)
	if err != nil {
		return err
	}

	// Checks if a user is in a group based on the group string
	if user.CheckRole(group) {
		return errors.New("User Already in Group " + group + "!")
	}

	// Open the "Users" bucket in the database
	db := h.db.rawdb.From("Users")

	// Assign the group to our target user
	user.SetRole(group)
	// Save the user changes in the database
	db.Update(&user)
	return nil
}

// ReadDemote The promote command runs using our commands array to get the promotion settings
func (h *PermissionsHandler) ReadDemote(commands []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	if len(commands) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Usage: demote <user> <group>")
		return
	}
	if len(m.Mentions) < 1 {
		s.ChannelMessageSend(m.ChannelID, "User must be mentioned")
		return
	}

	// Grab our target user id and group
	target := m.Mentions[0].ID
	group := commands[2]

	// Get the authors user object from the database
	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Check the group argument
	if group == "owner" {
		if !user.CheckRole("owner") {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" https://www.youtube.com/watch?v=7qnd-hdmgfk ")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to owner"
			return
		}
		s.ChannelMessageSend(m.ChannelID, "This group cannot be assigned through the promote command.")
		h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to owner"
		return

	}
	if group == "admin" {
		if !user.CheckRole("owner") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to admin"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to admin || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been set to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "smoderator" {

		if !user.CheckRole("admin") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to smoderator"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to smoderator || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been set to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "moderator" {

		if !user.CheckRole("smoderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run promote to moderator"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to moderator || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been set to the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been demoted to the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "builder" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to editor"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to editor || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been removed from the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "writer" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to agora"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to agora || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been removed from the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "scripter" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to streamer"
			return
		}
		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to streamer || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been removed from the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention()
		return

	}
	if group == "architect" {

		if !user.CheckRole("moderator") {
			s.ChannelMessageSend(m.ChannelID, "You do not have permission to assign this group")
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to recruiter"
			return
		}

		err = h.Demote(target, group)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			h.logchan <- "Permissions " + m.Author.Mention() + " attempted to run demote to recruiter || " +
				target + "||" + group + "||" + err.Error()
			return
		}
		s.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+" has been removed from the "+group+" group.")
		h.logchan <- "Permissions " + m.Mentions[0].Mention() + " has been removed from the " + group + " group by " + m.Author.Mention()
		return

	}
	s.ChannelMessageSend(m.ChannelID, group+" is not a valid group!")
	h.logchan <- "Permissions " + m.Author.Mention() + " attempted to demote " + m.Mentions[0].Mention() +
		" to " + group + " which does not exist"
	return

}

// Demote function
func (h *PermissionsHandler) Demote(userid string, group string) (err error) {

	// Open the "Users" bucket in the database
	db := h.db.rawdb.From("Users")

	// Get user from the database using the userid
	userobject := User{}
	err = db.One("ID", userid, &userobject)
	if err != nil {
		return err
	}

	if group == "smoderator" {
		userobject.SetRole("smoderator")
	}
	if group == "moderator" {
		userobject.RemoveRole("admin")
		userobject.RemoveRole("smoderator")
		userobject.SetRole("moderator")
	}
	if group == "builder" {
		userobject.RemoveRole("builder")
	}
	if group == "writer" {
		userobject.RemoveRole("writer")
	}

	if group == "scripter" {
		userobject.RemoveRole("scripter")
	}

	if group == "architect" {
		userobject.RemoveRole("architect")
	}

	err = db.DeleteStruct(&userobject)
	if err != nil {
		return err
	}
	err = db.Save(&userobject)
	if err != nil {
		return err
	}

	return nil
}



func (h *PermissionsHandler) CreatePermissionOverwrite(roleid string, permtype string, allow bool) (overwrite discordgo.PermissionOverwrite, err error) {

	if allow {
		overwrite = discordgo.PermissionOverwrite{ID: roleid, Type: permtype, Deny: 0, Allow: 1}
	} else {
		overwrite = discordgo.PermissionOverwrite{ID: roleid, Type: permtype, Deny: 1, Allow: 0}
	}

	return overwrite, nil
}

func (h *PermissionsHandler) CreatePermissionInt(roleperms RolePermissions ) (perm int){

	perm = 0
	if roleperms.CREATE_INSTANT_INVITE {
		perm = perm | 0x00000001
	}
	if roleperms.KICK_MEMBERS {
		perm = perm | 0x00000002
	}
	if roleperms.BAN_MEMBERS {
		perm = perm | 0x00000004
	}
	if roleperms.ADMINISTRATOR {
		perm = perm | 0x00000008
	}
	if roleperms.MANAGE_CHANNELS {
		perm = perm | 0x00000010
	}
	if roleperms.MANAGE_GUILD {
		perm = perm | 0x00000020
	}
	if roleperms.ADD_REACTIONS {
		perm = perm | 0x00000040
	}
	if roleperms.VIEW_AUDIT_LOG {
		perm = perm | 0x00000080
	}
	if roleperms.VIEW_CHANNEL {
		perm = perm | 0x00000400
	}
	if roleperms.SEND_MESSAGES {
		perm = perm | 0x00000800
	}
	if roleperms.SEND_TTS_MESSAGES {
		perm = perm | 0x00001000
	}
	if roleperms.MANAGE_MESSAGES {
		perm = perm | 0x00002000
	}
	if roleperms.EMBED_LINKS {
		perm = perm | 0x00004000
	}
	if roleperms.ATTACH_FILES {
		perm = perm | 0x00008000
	}
	if roleperms.READ_MESSAGE_HISTORY {
		perm = perm | 0x00010000
	}
	if roleperms.MENTION_EVERYONE {
		perm = perm | 0x00020000
	}
	if roleperms.USE_EXTERNAL_EMOJIS {
		perm = perm | 0x00040000
	}
	if roleperms.CONNECT {
		perm = perm | 0x00100000
	}
	if roleperms.SPEAK {
		perm = perm | 0x00200000
	}
	if roleperms.MUTE_MEMBERS {
		perm = perm | 0x00400000
	}
	if roleperms.DEAFEN_MEMBERS {
		perm = perm | 0x00800000
	}
	if roleperms.MOVE_MEMBERS {
		perm = perm | 0x01000000
	}
	if roleperms.USE_VAD {
		perm = perm | 0x02000000
	}
	if roleperms.CHANGE_NICKNAME {
		perm = perm | 0x04000000
	}
	if roleperms.MANAGE_NICKNAMES {
		perm = perm | 0x08000000
	}
	if roleperms.MANAGE_ROLES {
		perm = perm | 0x10000000
	}
	if roleperms.MANAGE_WEBHOOKS {
		perm = perm | 0x20000000
	}
	if roleperms.MANAGE_EMOJIS {
		perm = perm | 0x40000000
	}

	return perm
}


func (h *PermissionsHandler) CreateRole(name string, guildID string, hoist bool, mentionable bool, color int, perm int, s *discordgo.Session) (createdrole *discordgo.Role, err error){

	// Capitalize roles
	name = strings.Title(name)

	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return createdrole, err
	}

	if len(roles) >= 90 {
		return createdrole, errors.New("Maximum supported roles reached!")
	}

	for _, role := range roles {
		if role.Name == name {
			return createdrole, errors.New("Role already exists with name: " + name)
		}
	}

	createdrole, err = s.GuildRoleCreate(guildID)
	createdrole.Name = name

	createdrole, err = s.GuildRoleEdit(guildID, createdrole.ID, name, color, hoist, perm, mentionable)
	if err != nil {
		return createdrole, err
	}

	return createdrole, nil
}


func (h *PermissionsHandler) AddRoleToUser(rolename string, userID string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	// Capitalize roles
	rolename = strings.Title(rolename)

	// Get user from the database using the userid
	user, err := h.user.GetUser(userID, s, m.ChannelID)
	if err != nil {
		return err
	}

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		return err
	}

	// Checks if Role Exists in Guild
	roleID, err := getRoleIDByName(s, guildID, rolename)
	if err == nil {
		_ = s.GuildMemberRoleAdd(guildID, user.ID, roleID)
	}


	// Checks if a user is in a role based on the group string
	if !user.CheckRole(rolename) {
		// Open the "Users" bucket in the database
		db := h.db.rawdb.From("Users")

		// Assign the group to our target user
		user.SetRole(rolename)
		// Save the user changes in the database
		err = db.Update(&user)
		if err != nil {
			return err
		}
	}

	_ = s.GuildMemberRoleAdd(user.GuildID, user.ID, roleID)

	return nil
}

func (h *PermissionsHandler) RemoveRoleFromUser(rolename string, userID string, s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	// Capitalize Roles
	rolename = strings.Title(rolename)

	// Get user from the database using the userid
	user, err := h.user.GetUser(userID, s, m.ChannelID)
	if err != nil {
		return err
	}

	guildID, err := getGuildID(s, m.ChannelID)
	if err != nil {
		return err
	}

	// Checks if Role Exists in Guild
	roleID, err := getRoleIDByName(s, guildID, rolename)
	if err == nil {
		_ = s.GuildMemberRoleRemove(guildID, user.ID, roleID)
	}

	// Checks if a user is in a role based on the group string
	if user.CheckRole(rolename) {
		// Open the "Users" bucket in the database
		db := h.db.rawdb.From("Users")

		// Remove role from our target user
		user.RemoveRole(rolename)
		// Save the user changes in the database
		err = db.Update(&user)
		if err != nil {
			return err
		}
	}

	// Checks if Role Exists in Guild
	roleID, err = getRoleIDByName(s, user.GuildID, rolename)
	if err == nil {
		_ = s.GuildMemberRoleRemove(user.GuildID, user.ID, roleID)
	}

	return nil
}


func (h *PermissionsHandler) SyncRolesDB( userID string, guildID string, channelID string, s *discordgo.Session) (err error){

	// Get user from the database using the userid
	user, err := h.user.GetUser(userID, s, channelID)
	if err != nil {
		return err
	}

	userroles := user.Roles

	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return err
	}


	guildroles, err := s.GuildRoles(guildID)
	if err != nil {
		return err
	}

	for _, guildrole := range guildroles {
		for _, memberrole := range member.Roles {
			if memberrole == guildrole.ID {
				addtorolelist := true
				for _, userrole := range userroles {
					if userrole == guildrole.Name {
						addtorolelist = false
					}
				}
				if (addtorolelist){
					userroles = append(userroles, guildrole.Name)
				}
			}
		}
	}
	return nil
}

func (h *PermissionsHandler) SyncServerRoles( userID string, channelID string, s *discordgo.Session) (err error){

	// Get user from the database using the userid
	user, err := h.user.GetUser(userID, s, channelID)
	if err != nil {
		return err
	}

	for _, role := range user.Roles {
		time.Sleep(time.Duration(time.Second*1))
		_ = s.GuildMemberRoleAdd(user.GuildID, user.ID, role)
	}

	return nil
}


func (h *PermissionsHandler) DeleteRoleOnGuild(roleID string, guildID string, s *discordgo.Session) (err error){

	rooms, err := h.room.rooms.GetAllRooms()

	inUse := ""
	for _, room := range rooms {
		if room.TravelRoleID == roleID{
			inUse = inUse + room.ID + " - " + room.Name +"\n"
		} else {
			room.RoleIDs = RemoveStringFromSlice(room.RoleIDs, roleID)
			err = h.room.rooms.SaveRoomToDB(room)
			if err != nil {
				return err
			}
		}
	}


	if inUse != "" {
		return errors.New("Cannot delete role configured as travel ID, it is currently in " +
									"use by the following channels: \n```\n"+inUse+"\n```\n")
	}

	return s.GuildRoleDelete(guildID, roleID)
}


func (h *PermissionsHandler) TranslateRoleID(roleID string, guildID string, s *discordgo.Session) (rolename string, err error) {

	rolename, err = getRoleNameByID(roleID, guildID, s)
	if err != nil {
		return "", err
	}

	return rolename, nil
}