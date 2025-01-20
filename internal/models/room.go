package models

type Room struct {
	ID      string
	Players []Player
	Leader  *Player
	
}
