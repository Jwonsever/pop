package pop

import (
	"io"

	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/pop/columns"
)

type Dialect interface {
	Name() string
	URL() string
	MigrationURL() string
	Details() *ConnectionDetails
	TranslateSQL(string) string
	Create(Store, *Model, columns.Columns) error
	Update(Store, *Model, columns.Columns) error
	Destroy(Store, *Model) error
	SelectOne(Store, *Model, Query) error
	SelectMany(Store, *Model, Query) error
	CreateDB() error
	DropDB() error
	DumpSchema(io.Writer) error
	LoadSchema(io.Reader) error
	FizzTranslator() fizz.Translator
	Lock(func() error) error
	TruncateAll(*Connection) error
	Quote(key string) string
}

type afterOpenable interface {
	AfterOpen(*Connection) error
}
