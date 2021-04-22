package mssql

import (
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Update: resourceDatabaseUpdate,
		Delete: resourceDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, m interface{}) error {
	db := m.(*sql.DB)
	name := d.Get("name").(string)
	owner := d.Get("owner").(string)
	_, err := db.Query(fmt.Sprintf(`CREATE DATABASE %s`, name))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("USE %s; CREATE USER %s FOR LOGIN %s; ALTER ROLE db_owner ADD MEMBER %s", name, owner, owner, owner))
	if err != nil {
		return err
	}

	row, err := checkTable(db, name)
	if err != nil {
		return err
	}
	d.SetId(row.name)

	return err
}

type DatabaseSchemaRow struct {
	name string
}

func resourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	db := m.(*sql.DB)
	name := d.Id()
	row, err := checkTable(db, name)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	if err = d.Set("name", row.name); err != nil {
		return err
	}

	return nil
}

func checkTable(db *sql.DB, name string) (*DatabaseSchemaRow, error) {
	var row DatabaseSchemaRow
	err := db.QueryRow(fmt.Sprintf("SELECT name FROM sys.databases where name = '%s'", name)).Scan(&row.name)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func resourceDatabaseUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, m interface{}) error {
	db := m.(*sql.DB)
	name := d.Id()
	_, err := db.Exec(fmt.Sprintf("ALTER DATABASE %s SET SINGLE_USER WITH ROLLBACK IMMEDIATE; DROP DATABASE %s;", name, name))
	if err != nil {
		return err
	}

	return err
}
