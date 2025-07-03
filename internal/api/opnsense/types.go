package opnsense

type HostOverride struct {
	UUID        string `json:"uuid"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	Server      string `json:"server"`
	Description string `json:"description"`
	Enabled     string `json:"enabled"`
}

type SearchResponse struct {
	Status string         `json:"status"`
	Rows   []HostOverride `json:"rows"`
}

type Response struct {
	Status string `json:"status"`
	Result string `json:"result"`
}
