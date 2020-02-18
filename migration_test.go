package migrate

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

func ExampleNewMigration() {
	// Create a dummy migration body, this is coming from the source usually.
	body := ioutil.NopCloser(strings.NewReader("dumy migration that creates users table"))

	// Create a new Migration that represents version 1486686016.
	// Once this migration has been applied to the database, the new
	// migration version will be 1486689359.
	migr, err := NewMigration(body, "create_users_table", source.Up, "create_users_table.up.sql")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(migr.LogString())
	// Output:
	// create_users_table/u create_users_table.up.sql
}
