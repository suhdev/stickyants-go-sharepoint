package sharepoint

type SPRequest struct {
	selections []string
	filters    []string
	orders     []string
}

func (r *SPRequest) Select(prop string) *SPRequest {
	append(r.selections, prop)
	return r
}

func (r *SPRequest) Filter(filter string) *SPRequest {
	append(r.fitlers, filter)
	return r
}

func (r *SPRequest) OrderBy(field string)
