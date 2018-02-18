package main

// User struct
type User struct {
	ID string `storm:"id"` // primary key

	Perms 		[]uint64 // Internal Permissions - NOT Discord Roles
	Roles		[]string
	/*Owner      bool `storm:"index"`	1000
	Admin      bool `storm:"index"`		90
	SModerator bool `storm:"index"`		80
	Moderator  bool `storm:"index"`		70
	Builder     bool `storm:"index"`	60
	Writer      bool `storm:"index"`	50
	Scripter   bool `storm:"index"`		40
	Architect  bool `storm:"index"`		30
	Player    bool `storm:"index"`		10
	*/
	GuildID		string `storm:"index"` // GuildID of the users current guild
	RoomID    string `storm:"index"` // ChannelID of the users current room
	ItemsMap	[]string	// An ID pointing to the

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
		u.JoinRole(role)
	}
}

// RemoveRole function
func (u *User) RemoveRole(role string) {

	switch role {

	case "owner":
		SetBit(&u.Perms, 1000)

	case "admin":
		SetBit(&u.Perms, 90)

	case "smoderator":
		SetBit(&u.Perms, 80)

	case "moderator":
		SetBit(&u.Perms, 70)

	case "builder":
		SetBit(&u.Perms, 60)

	case "writer":
		SetBit(&u.Perms, 50)

	case "scripter":
		SetBit(&u.Perms, 40)

	case "architect":
		SetBit(&u.Perms, 30)

	case "player":
		SetBit(&u.Perms, 10)

	default:
		u.LeaveRole(role)
	}
}

// CheckRole function
func (u *User) CheckRole(role string) bool {

	switch role {

	case "owner":
		return IsBitSet(&u.Perms, 1000)

	case "admin":
		return IsBitSet(&u.Perms, 90)

	case "smoderator":
		return IsBitSet(&u.Perms, 80)

	case "moderator":
		return IsBitSet(&u.Perms, 70)

	case "builder":
		return IsBitSet(&u.Perms, 60)

	case "writer":
		return IsBitSet(&u.Perms, 50)

	case "scripter":
		return IsBitSet(&u.Perms, 40)

	case "architect":
		return IsBitSet(&u.Perms, 30)

	case "player":
		return IsBitSet(&u.Perms, 10)

	default:
		return u.CheckDiscordRole(role)
	}
}

func (u *User) CheckDiscordRole(rolename string) bool {

	for _, role := range u.Roles {

		if role == rolename{
			return true
		}
	}
	return false
}

func (u *User) JoinRole(rolename string) {
	if u.CheckDiscordRole(rolename){
		return
	}

	u.Roles = append(u.Roles, rolename)

}


func (u *User) LeaveRole(rolename string) {
	if !u.CheckDiscordRole(rolename){
		return
	}

	u.Roles = RemoveStringFromSlice(u.Roles, rolename)
}