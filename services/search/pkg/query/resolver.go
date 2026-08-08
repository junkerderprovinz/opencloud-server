package query

import (
	"reflect"
	"strings"
	"sync"

	"github.com/opencloud-eu/opencloud/services/search/pkg/mapping"
	"github.com/opencloud-eu/opencloud/services/search/pkg/search"
)

// aliases are KQL spellings the derived index can't produce (fields are plural).
var aliases = map[string]string{
	"tag":      "Tags",
	"favorite": "Favorites",
}

// fieldIndex maps a lowercased KQL key to the real field name: derived once from
// the resource struct, overlaid with the explicit aliases.
var fieldIndex = sync.OnceValue(func() map[string]string {
	idx := mapping.FieldNameIndex(
		reflect.TypeFor[search.Resource](),
		search.Resource{}.SearchFieldOverrides(),
	)
	for k, v := range aliases {
		idx[k] = v
	}
	return idx
})

// ResolveField maps a KQL key to the index field name: empty -> Name, a known
// key (case-insensitive) -> its field, anything else unchanged.
func ResolveField(name string) string {
	if name == "" {
		return "Name"
	}
	if v, ok := fieldIndex()[strings.ToLower(name)]; ok {
		return v
	}
	return name
}

// geopointFields maps a lowercased KQL key to its indexed geopoint sibling field,
// derived from the TypeGeopoint entries in the resource field overrides (e.g.
// "location" -> "location_geopoint", "journey.start" -> "journey.start_geopoint").
var geopointFields = sync.OnceValue(func() map[string]string {
	out := map[string]string{}
	for key, opts := range (search.Resource{}).SearchFieldOverrides() {
		if opts.Type == mapping.TypeGeopoint {
			out[strings.ToLower(key)] = key + mapping.GeopointSuffix
		}
	}
	return out
})

// ResolveGeoField maps a KQL key to its indexed geopoint field name. ok is false
// when the key is not a geopoint field, so callers can reject geo predicates on
// non-geo fields.
func ResolveGeoField(name string) (string, bool) {
	f, ok := geopointFields()[strings.ToLower(name)]
	return f, ok
}
