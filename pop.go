package pop

import "strings"

// AvailableDialects lists the available database dialects
var AvailableDialects []string

var dialectSynonyms = make(map[string]string)

// map of dialect specific url parsers
var urlParser = make(map[string]func(*ConnectionDetails) error)

// map of dialect specific connection details finalizers
var finalizer = make(map[string]func(*ConnectionDetails))

// map of connection creators
var newConnection = make(map[string]func(*ConnectionDetails) (dialect, error))

// EagerMode type for all eager modes supported in pop.
type EagerMode int8

const (
	eagerModeNil EagerMode = iota
	// EagerDefault is the current implementation, the default
	// behavior of pop. This one introduce N+1 problem and will be used as
	// default value for backward compatibility.
	EagerDefault

	// EagerPreload mode works similar to Preload mode used in ActiveRecord.
	// Avoid N+1 problem by reducing the number of hits to the database but
	// increase memory use to process and link associations to parent.
	EagerPreload

	// EagerInclude This mode works similar to Include mode used in rails ActiveRecord.
	// Use Left Join clauses to load associations.
	EagerInclude
)

// default loading Association Strategy definition.
var loadingAssociationsStrategy = EagerDefault

// SetEagerMode changes overall mode when eager loading.
// this will change the default loading associations strategy for all Eager queries.
// This should be used once, when setting up pop connection.
// func SetEagerMode(eagerMode EagerMode) {
// 	loadingAssociationsStrategy = eagerMode
// }

// DialectSupported checks support for the given database dialect
func DialectSupported(d string) bool {
	for _, ad := range AvailableDialects {
		if ad == d {
			return true
		}
	}
	return false
}

func normalizeSynonyms(dialect string) string {
	d := strings.ToLower(dialect)
	if syn, ok := dialectSynonyms[d]; ok {
		d = syn
	}
	return d
}
