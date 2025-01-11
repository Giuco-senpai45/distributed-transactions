package models

import (
	"strings"
)

type Path struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (p *Path) GetParent() string {
	parts := strings.Split(p.Path, ".")
	if len(parts) == 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

func (p *Path) IsDescendantOf(ancestor string) bool {
	return strings.HasPrefix(p.Path, ancestor+".")
}

func (p *Path) IsAncestorOf(descendant string) bool {
	return strings.HasPrefix(descendant, p.Path+".")
}
