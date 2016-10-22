package proto

type Server struct {
	Id         string         `json:"id,omitempty"`
	AssetId    uint64         `json:"assetId,omitempty"`
	Datacenter string         `json:"datacenter,omitempty"`
	Location   string         `json:"location,omitempty"`
	Name       string         `json:"name,omitempty"`
	IsOnline   bool           `json:"isOnline,omitempty"`
	IsDeleted  bool           `json:"isDeleted,omitempty"`
	Details    *ServerDetails `json:"details,omitempty"`
}

type ServerDetails struct {
	DetailsCreation
	/*
		Nodes     []string `json:"nodes,omitempty"`
	*/
}

type ServerFilter struct {
	IsOnline   bool   `json:"isOnline,omitempty"`
	NotOnline  bool   `json:"notOnline,omitempty"`
	Deleted    bool   `json:"Deleted,omitempty"`
	NotDeleted bool   `json:"notDeleted,omitempty"`
	Datacenter string `json:"datacenter,omitempty"`
	Name       string `json:"name,omitempty"`
	AssetId    uint64 `json:"assetId,omitempty"`
}

func (p *Server) DeepCompare(a *Server) bool {
	if p.Id != a.Id || p.AssetId != a.AssetId || p.Datacenter != a.Datacenter ||
		p.Location != a.Location || p.Name != a.Name || p.IsOnline != a.IsOnline ||
		p.IsDeleted != a.IsDeleted {
		return false
	}
	return true
}

func NewServerRequest() Request {
	return Request{
		Flags:  &Flags{},
		Server: &Server{},
	}
}

func NewServerFilter() Request {
	return Request{
		Filter: &Filter{
			Server: &ServerFilter{},
		},
	}
}

func NewServerResult() Result {
	return Result{
		Errors:  &[]string{},
		Servers: &[]Server{},
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix