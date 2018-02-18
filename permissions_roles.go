package main

// OwnerRole function
func OwnerRole(u *User) {
	SetBit(&u.Perms, 1000) // Owner
	SetBit(&u.Perms, 90) // Admin
	SetBit(&u.Perms, 80) // Smoderator
	SetBit(&u.Perms, 70) // Moderator
	SetBit(&u.Perms, 60) // Builder
	SetBit(&u.Perms, 50) // Writer
	SetBit(&u.Perms, 40) // Scripter
	SetBit(&u.Perms, 30) // Architect
	SetBit(&u.Perms, 10) // Player

}

// AdminRole function
func AdminRole(u *User) {
	SetBit(&u.Perms, 90) // Admin
	SetBit(&u.Perms, 80) // Smoderator
	SetBit(&u.Perms, 70) // Moderator
	SetBit(&u.Perms, 60) // Builder
	SetBit(&u.Perms, 50) // Writer
	SetBit(&u.Perms, 40) // Scripter
	SetBit(&u.Perms, 10) // Player

}

// SModeratorRole function
func SModeratorRole(u *User) {
	SetBit(&u.Perms, 80) // Smoderator
	SetBit(&u.Perms, 70) // Moderator
	SetBit(&u.Perms, 60) // Builder
	SetBit(&u.Perms, 50) // Writer
	SetBit(&u.Perms, 40) // Scripter
	SetBit(&u.Perms, 10) // Player

}

// ModeratorRole function
func ModeratorRole(u *User) {
	SetBit(&u.Perms, 70) // Moderator
}

// BuilderRole function
func BuilderRole(u *User) {
	SetBit(&u.Perms, 60) // Builder
}

// WriterRole function
func WriterRole(u *User) {
	SetBit(&u.Perms, 50) // Writer
}

// StreamerRole function
func ScripterRole(u *User) {
	SetBit(&u.Perms, 40) // Scripter
}

// RecruiterRole function
func ArchitectRole(u *User) {
	SetBit(&u.Perms, 30) // Scripter
}

// ClearRoles function
func ClearRoles(u *User) {
	ClearBit(&u.Perms, 90) // Admin
	ClearBit(&u.Perms, 80) // Smoderator
	ClearBit(&u.Perms, 70) // Moderator
	ClearBit(&u.Perms, 60) // Builder
	ClearBit(&u.Perms, 50) // Writer
	ClearBit(&u.Perms, 40) // Scripter
	ClearBit(&u.Perms, 30) // Architect
	ClearBit(&u.Perms, 10) // Player
}

// CitizenRole function
func PlayerRole(u *User) {
	SetBit(&u.Perms, 10) // Player
}
