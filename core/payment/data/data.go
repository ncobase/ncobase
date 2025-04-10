package data

import (
	"github.com/ncobase/ncore/pkg/config"
	"github.com/ncobase/ncore/pkg/data"
)

// Data .
type Data struct {
	*data.Data
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data: d,
	}, cleanup, nil
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}
	return errs
}

/* Example usage:

// Write operations
db := d.GetDB()
_, err = db.Exec("INSERT INTO users (name) VALUES (?)", "test")

// Read operations
dbRead, err := d.GetDBRead()
if err != nil {
    // handle error
}
rows, err := dbRead.Query("SELECT * FROM users")

// Write transaction
err := d.WithTx(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    _, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "test")
    return err
})

// Read-only transaction
err := d.WithTxRead(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    rows, err := tx.Query("SELECT * FROM users")
    return err
})
*/
