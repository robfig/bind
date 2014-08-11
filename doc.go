/*
Package bind converts between form encoding and Go values.

It comes with binders for all values, time.Time, arbitrary structs, and
slices. In particular, binding functions are provided for the following types:

 - bool
 - float32, float64
 - int, int8, int16, int32, int64
 - uint, uint8, uint16, uint32, uint64
 - string
 - struct
 - a pointer to any supported type
 - a slice of any supported type
 - time.Time
 - uploaded files (as io.Reader, io.ReadSeeker, *os.File, []byte, *multipart.FileHeader)

Callers may also hook into the process and provide a custom binding function.

Example

This example binds data from embedded URL arguments, the query string, and a
posted form.

 POST /accounts/:accountId/users/?op=UPDATE

 <form>
  <input name="user.Id">
  <input name="user.Name">
  <input name="user.Phones[0].Label">
  <input name="user.Phones[0].Number">
  <input name="user.Phones[1].Label">
  <input name="user.Phones[1].Number">
  <input name="user.Labels[]">
  <input name="user.Labels[]">
 </form>

 type Phone struct { Label, Number string }
 type User struct {
   Id     uint32
   Phones []Phone
   Labels []string
 }

 var (
   params = mux.Vars(req) // embedded URL args
   id     uint32
   op     string
   user   User
 )
 handleErrors(
   bind.Map(params).Field(&id, "accountId"),
   bind.Request(req).Field(&op, "op")
   bind.Request(req).Field(&user, "user"),
 )


Booleans

Booleans are converted to Go by comparing against the following strings:

  TRUE: "true", "1",  "on"
  FALSE: "false", "0", ""

The "on" / "" syntax is supported as the default behavior for HTML checkboxes.


Date Time

The SQL standard time formats [“2006-01-02”, “2006-01-02 15:04”] are recognized
by the default datetime binder.

More may be added by the application to the TimeFormats variable, like this:

 func init() {
	 bind.TimeFormats = append(bind.TimeFormats, "01/02/2006")
 }

File Uploads

File uploads may be bound to any of the following types:

 - *os.File
 - []byte
 - io.Reader
 - io.ReadSeeker

This is a wrapper around the upload handling provided by Go’s multipart
package. The bytes stay in memory unless they exceed a threshold (10MB by
default), in which case they are written to a temp file.

Note: Binding a file upload to os.File requires Revel to write it to a temp file
(if it wasn’t already), making it less efficient than the other types.

Slices

Both indexed and unindexed slices are supported.

These two forms are bound as unordered slices:

 <form>
  <input name="ids">
  <input name="ids">
  <input name="ids">
 </form>

 <form>
  <input name="ids[]">
  <input name="ids[]">
  <input name="ids[]">
 </form>

This is bound as an ordered slice:

 <form>
  <input name="ids[0]">
  <input name="ids[1]">
  <input name="ids[2]">
 </form>

The two forms may be combined, with unindexed elements filling any gaps between
indexed elements.

 <form>
  <input name="ids[]">
  <input name="ids[]">
  <input name="ids[5]">
 </form>

Note that if the slice element is a struct, it must use the indexed notation.

Structs

Structs are bound using a dot notation.  For example:

 <form>
  <input name="user.Name">
  <input name="user.Phones[0].Label">
  <input name="user.Phones[0].Number">
  <input name="user.Phones[1].Label">
  <input name="user.Phones[1].Number">
 </form>

Struct fields must be exported to be bound.

Additionally, all params may be bound as members of a struct, rather than
extracting a single field.

 <form>
  <input name="Name">
  <input name="Phones[0].Label">
  <input name="Phones[0].Number">
  <input name="Phones[1].Label">
  <input name="Phones[1].Number">
 </form>

 var user User
 err := bind.Request(req).All(&user)

*/
package bind
