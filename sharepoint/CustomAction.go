package sharepoint

type CustomActionElements struct {
	Items []*CustomAction `json:"Items"`
}

type ActionRights struct {
	High string `json:"High"`
	Low  string `json:"Low"`
}

type CustomAction struct {
	ClientSideComponentId         string        `json:"ClientSideComponentId"`
	ClientSideComponentProperties string        `json:"ClientSideComponentProperties"`
	Id                            string        `json:"Id"`
	ImageUrl                      string        `json:"ImageUrl"`
	Location                      string        `json:"Location"`
	RegistrationId                string        `json:"RegistrationId"`
	RegistrationType              int           `json:"RegistrationType"`
	RequireSiteAdministrator      bool          `json:"RequireSiteAdministrator"`
	Rights                        *ActionRights `json:"Rights"`
	Title                         string        `json:"Title"`
	UrlAction                     string        `json:"UrlAction"`
}
