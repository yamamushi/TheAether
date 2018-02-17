package main

// User struct
type User struct {
	ID string `storm:"id"` // primary key

	Owner      bool `storm:"index"`
	Admin      bool `storm:"index"`
	SModerator bool `storm:"index"`
	Moderator  bool `storm:"index"`
	Builder     bool `storm:"index"`
	Writer      bool `storm:"index"`
	Scripter   bool `storm:"index"`
	Architect  bool `storm:"index"`
	Player    bool `storm:"index"`
	GuildID		string `storm:"index"` // GuildID of the users current guild
	RoomID    string `storm:"index"` // ChannelID of the users current room

}

// Init function
func (u *User) Init() {
	ClearRoles(u)
	PlayerRole(u)
}

// SetRole function
func (u *User) SetRole(role string) {

	switch role {

	case "owner":
		OwnerRole(u)

	case "admin":
		AdminRole(u)

	case "smoderator":
		SModeratorRole(u)

	case "moderator":
		ModeratorRole(u)

	case "builder":
		BuilderRole(u)

	case "writer":
		WriterRole(u)

	case "scripter":
		ScripterRole(u)

	case "architect":
		ArchitectRole(u)

	case "player":
		PlayerRole(u)

	case "clear":
		ClearRoles(u)

	default:
		return
	}
}

// RemoveRole function
func (u *User) RemoveRole(role string) {

	switch role {

	case "owner":
		u.Owner = false

	case "admin":
		u.Admin = false

	case "smoderator":
		u.SModerator = false

	case "moderator":
		u.Moderator = false

	case "builder":
		u.Builder = false

	case "writer":
		u.Writer = false

	case "scripter":
		u.Scripter = false

	case "architect":
		u.Architect = false

	case "player":
		u.Player = false

	}
}

// CheckRole function
func (u *User) CheckRole(role string) bool {

	switch role {

	case "owner":
		return u.Owner

	case "admin":
		return u.Admin

	case "smoderator":
		return u.SModerator

	case "moderator":
		return u.Moderator

	case "builder":
		return u.Builder

	case "writer":
		return u.Writer

	case "scripter":
		return u.Scripter

	case "architect":
		return u.Architect

	case "player":
		return u.Player
	}

	return false
}
