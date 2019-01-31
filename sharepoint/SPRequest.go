package sharepoint

type SPRequest struct {
	selections []string
	expand     []string
	filters    []string
	orders     []string
}

func (r *SPRequest) Select(prop string) *SPRequest {
	r.selections = append(r.selections, prop)
	return r
}

func (r *SPRequest) Filter(filter string) *SPRequest {
	r.filters = append(r.filters, filter)
	return r
}

func (r *SPRequest) OrderBy(field string) *SPRequest {

}
