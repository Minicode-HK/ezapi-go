package cms

import (
    _ "simple_backend_go/route"
)

// Register your models here
// Leave empty to include *ALL* models automatically
var Modules []interface{} = []interface{}{
	// Add more models as needed
}

var Users = map[string]string{
	"superadmin": "superadmin",
	// Add more users as needed
}