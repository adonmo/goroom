package main

import (
	"adonmo.com/goroom/room"
)

//InitializeRoom Initialize Room
func InitializeRoom(initializer room.Initializer, fallbackToDestructiveMigration bool) (errList []error) {

	identityHash, err := initializer.CalculateIdentityHash()
	if err != nil {
		return append(errList, err)
	}

	shouldRetryAfterDestruction, err := initializer.Init(identityHash)
	if err != nil && shouldRetryAfterDestruction && fallbackToDestructiveMigration {
		if err = initializer.PerformDBCleanUp(); err == nil {
			_, err = initializer.Init(identityHash)
		}
	}

	return append(errList, err)
}
