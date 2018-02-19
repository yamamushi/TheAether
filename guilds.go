package main

import "sync"

type GuildsManager struct {

	db          *DBHandler
	querylocker sync.RWMutex

}


type GuildRecord struct {

	ID string `storm:"id"` // primary key

}