//Package types define all common types
package types

import "github.com/lonnng/nanoserver/db"

//Closer the close handler
type Closer func()

//UserMeta user's meta info
type UserMeta struct {
	Role  int    `redis:"role"`
	Uid   int64  `redis:"uid"`
	AppID string `redis:"appid"`
}

//RoleIsSnake check the user's role is admin or not
func (um *UserMeta) RoleIsAdmin() bool {
	return um.Role == db.RoleTypeAdmin
}

//UID return the user's uid
func (um *UserMeta) UID() int64 {
	return um.Uid
}
