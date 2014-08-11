/*
Package bind provides conversions between form encoding and Go values.

Example

 var (
   binder = bind.Request(req)
   id   uint32
   user User
 )
 handleErrors(
   bind.Map(params).Field("id", &id),
   bind.Request(req).Field("user", &user),
 )

// todo:

 {
  "id"
  "name"
  "xxx"
 }

 bind.Map(..).All(&user)

*/
package bind
