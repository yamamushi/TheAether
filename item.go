package main


// ItemType struct
type ItemType struct {

	ID string `storm:"id"` // primary key

	ItemType	string	`storm:"index"`
	OwnerID		string	`storm:"index"`
	Description string
	Weight		float64
	Durability	float64
	Status 		[]string

}

