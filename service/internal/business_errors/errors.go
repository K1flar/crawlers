package business_errors

var (
	InvalidQuery      = New("invalid_query")
	UnavailableSource = New("unavailable_source")
	EntityNotFound    = New("entity_not_found")
)
