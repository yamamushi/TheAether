package main

// Permissions struct
type Permissions struct{}

// CommandPermissions struct
type CommandPermissions struct {
	ID      string `storm:"id"`
	command string `storm:"index"`
	channel string `storm:"index"`
	groups  []string
}

type RolePermissions struct {

	CREATE_INSTANT_INVITE 		bool
	KICK_MEMBERS				bool
	BAN_MEMBERS 				bool
	ADMINISTRATOR 				bool
	MANAGE_CHANNELS 			bool
	MANAGE_GUILD 				bool
	ADD_REACTIONS				bool
	VIEW_AUDIT_LOG				bool
	VIEW_CHANNEL				bool
	SEND_MESSAGES				bool
	SEND_TTS_MESSAGES			bool
	MANAGE_MESSAGES 			bool
	EMBED_LINKS					bool
	ATTACH_FILES				bool
	READ_MESSAGE_HISTORY		bool
	MENTION_EVERYONE			bool
	USE_EXTERNAL_EMOJIS			bool
	CONNECT						bool
	SPEAK						bool
	MUTE_MEMBERS				bool
	DEAFEN_MEMBERS				bool
	MOVE_MEMBERS				bool
	USE_VAD						bool
	CHANGE_NICKNAME				bool
	MANAGE_NICKNAMES			bool
	MANAGE_ROLES 				bool
	MANAGE_WEBHOOKS 			bool
	MANAGE_EMOJIS 				bool

}