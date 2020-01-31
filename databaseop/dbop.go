package databaseop

import (
	"github.com/matoous/go-nanoid"
	"log"
)

type DBClient interface {
	connNCheck(dbCliOption interface{}) error
	itemCreate(inputdata interface{}) error
	itemUpdate(filter1 interface{}, change1 interface{}) error
	itemDelete(filter1 interface{}) error
	itemRead(filter1 interface{}) (UserData, error)
}

func GetNanoID() (string, error) {
	id, err := gonanoid.Nanoid(4)
	if err != nil {
		log.Fatalln("Failed to generate nanoid!")
	}
	return id, err
}
