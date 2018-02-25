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

// RolePermissions struct
type RolePermissions struct {
	CreateInstantInvite bool
	KickMembers         bool
	BanMembers          bool
	Administrator       bool
	ManageChannels      bool
	ManageGuild         bool
	AddReactions        bool
	ViewAuditLog        bool
	ViewChannel         bool
	SendMessages        bool
	SendTTSMessages     bool
	ManageMessages      bool
	EmbedLinks          bool
	AttachFiles         bool
	ReadMessageHistory  bool
	MentionEveryone     bool
	UseExternalEmojis   bool
	Connect             bool
	Speak               bool
	MuteMembers         bool
	DeafenMEmbers       bool
	MoveMembers         bool
	UseVAD              bool
	ChangeNickname      bool
	ManageNicknames     bool
	ManageRoles         bool
	ManageWebhooks      bool
	ManageEmojis        bool
}
