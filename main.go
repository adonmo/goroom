package goroom

import (
	"adonmo.com/goroom/room"
)

//InitializeRoom Initialize Room
func InitializeRoom(initializer room.Initializer, fallbackToDestructiveMigration bool) error {

	identityHash, err := initializer.CalculateIdentityHash()
	if err != nil {
		return err
	}

	shouldRetryAfterDestruction, err := initializer.Init(identityHash)
	if err != nil && shouldRetryAfterDestruction && fallbackToDestructiveMigration {
		if err = initializer.PerformDBCleanUp(); err == nil {
			_, err = initializer.Init(identityHash)
		}
	}

	return err
}
