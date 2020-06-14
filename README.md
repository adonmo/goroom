# goroom

### Brief
The primary function of Room is to ease version management and migration of embedded Data Stores.
Embedded Data Stores are databases that are created by apps on the edge devices and are tightly coupled with the
version of app that creates and manages them.

### Features
* Using Room a developer can ensure that as they deliver updates to App and underlying associated Data Store they can
have a smooth transition of Data Store on the edge device without risk of data loss.  
* A typical way to handle upgrades would be to create the data store from scratch again. This is not desirable if the
data stored on the edge device is things like valuable insights/events recorded by the device and pending sync/upload to the
server. Since data collection is a major use case of edge devices I think a version manager like Room is necessary.

### Android Room
Room is inspired by its [namesake](https://developer.android.com/training/data-storage/room) in Android World which does the same thing but at a deeper level by even providing the ORM.
The Room presented here is agnostic to data stores and provides flexibility to the developer on how they [signal](https://github.com/gamble09/groom/blob/master/orm/orm.go#L31) and [handle schema changes](https://github.com/gamble09/groom/blob/master/orm/orm.go#L36).

### Gotchas
* It is purely a utility that serves the minimal purpose of carrying out migrations and verifying that DB is upto the version expected by the app currently.  
* A lot of power is still in the developers hands as they have the freedom to execute any operations on the DB themselves.
* Doing stuff like deleting/updating Room's metadata tables is a big No-No :). Plz...

### Sample
For understanding on how the migration and versioning works check [examples](https://github.com/gamble09/groom/tree/master/example).  

### How to Run Sample
To run the example which also serves as an integration test with GORM go to the examples folder and run
```sh
go test -v ./...
```
