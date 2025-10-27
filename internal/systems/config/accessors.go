package config

import "image/color"

// Theme accessors
func (r *RootConfig) GetCursorColor() color.Color {
	if r == nil {
		return color.RGBA{200, 200, 200, 200}
	}
	rgb := r.Theme.CursorColor
	a := r.Theme.CursorOpacity
	return color.RGBA{rgb[0], rgb[1], rgb[2], a}
}

func (r *RootConfig) GetCursorShape() string {
	if r == nil || r.Theme.CursorShape == "" {
		return "block"
	}
	return r.Theme.CursorShape
}

func (r *RootConfig) GetCursorBlink() bool {
	if r == nil {
		return false
	}
	return r.Theme.CursorBlink
}

func (r *RootConfig) GetFontName() string {
	if r == nil || r.Renderer.FontName == "" {
		return "basic"
	}
	return r.Renderer.FontName
}

func (r *RootConfig) GetShell() string {
	if r == nil || r.System.DefaultShell == "" {
		return "/bin/bash"
	}
	return r.System.DefaultShell
}

