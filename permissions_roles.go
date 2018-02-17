package main

// OwnerRole function
func OwnerRole(u *User) {
	u.Owner = true
	u.Admin = true
	u.SModerator = true
	u.Moderator = true
	u.Builder = true
	u.Writer = true
	u.Scripter = true
	u.Player = true
}

// AdminRole function
func AdminRole(u *User) {
	u.Admin = true
	u.SModerator = true
	u.Moderator = true
	u.Builder = true
	u.Writer = true
	u.Scripter = true
	u.Player = true
}

// SModeratorRole function
func SModeratorRole(u *User) {
	u.SModerator = true
	u.Moderator = true
	u.Builder = true
	u.Writer = true
	u.Scripter = true
	u.Player = true
}

// ModeratorRole function
func ModeratorRole(u *User) {
	u.Moderator = true
}

// BuilderRole function
func BuilderRole(u *User) {
	u.Builder = true
}

// WriterRole function
func WriterRole(u *User) {
	u.Writer = true
}

// StreamerRole function
func ScripterRole(u *User) {
	u.Scripter = true
}

// RecruiterRole function
func ArchitectRole(u *User) {
	u.Architect = true
}

// ClearRoles function
func ClearRoles(u *User) {
	u.Owner = false
	u.Admin = false
	u.SModerator = false
	u.Moderator = false
	u.Builder = false
	u.Writer = false
	u.Scripter = false
	u.Architect = false
}

// CitizenRole function
func PlayerRole(u *User) {
	u.Player = true
}
