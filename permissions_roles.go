package main

// OwnerRole function
func OwnerRole(u *User) {
	SetBit(&u.Perms, 60) // Owner
	SetBit(&u.Perms, 59) // Admin
	SetBit(&u.Perms, 58) // Smoderator
	SetBit(&u.Perms, 57) // Moderator
	SetBit(&u.Perms, 56) // Builder
	SetBit(&u.Perms, 55) // Writer
	SetBit(&u.Perms, 54) // Scripter
	SetBit(&u.Perms, 53) // Architect
	SetBit(&u.Perms, 10) // Player

}

// AdminRole function
func AdminRole(u *User) {
	SetBit(&u.Perms, 59) // Admin
	SetBit(&u.Perms, 58) // Smoderator
	SetBit(&u.Perms, 57) // Moderator
	SetBit(&u.Perms, 56) // Builder
	SetBit(&u.Perms, 55) // Writer
	SetBit(&u.Perms, 54) // Scripter
	SetBit(&u.Perms, 10) // Player

}

// SModeratorRole function
func SModeratorRole(u *User) {
	SetBit(&u.Perms, 58) // Smoderator
	SetBit(&u.Perms, 57) // Moderator
	SetBit(&u.Perms, 56) // Builder
	SetBit(&u.Perms, 55) // Writer
	SetBit(&u.Perms, 54) // Scripter
	SetBit(&u.Perms, 10) // Player

}

// ModeratorRole function
func ModeratorRole(u *User) {
	SetBit(&u.Perms, 57) // Moderator
}

// BuilderRole function
func BuilderRole(u *User) {
	SetBit(&u.Perms, 56) // Builder
}

// WriterRole function
func WriterRole(u *User) {
	SetBit(&u.Perms, 55) // Writer
}

// ScripterRole function
func ScripterRole(u *User) {
	SetBit(&u.Perms, 54) // Scripter
}

// ArchitectRole function
func ArchitectRole(u *User) {
	SetBit(&u.Perms, 53) // Architect
}

// ClearRoles function
func ClearRoles(u *User) {
	ClearBit(&u.Perms, 59) // Admin
	ClearBit(&u.Perms, 58) // Smoderator
	ClearBit(&u.Perms, 57) // Moderator
	ClearBit(&u.Perms, 56) // Builder
	ClearBit(&u.Perms, 55) // Writer
	ClearBit(&u.Perms, 54) // Scripter
	ClearBit(&u.Perms, 53) // Architect
	ClearBit(&u.Perms, 10) // Player
}

// PlayerRole function
func PlayerRole(u *User) {
	SetBit(&u.Perms, 10) // Player
}
