package config

// RootConfig defines persistent user-editable settings.
type RootConfig struct {
	Version  string       `json:"version"`
	Renderer RenderConfig `json:"renderer"`
	Theme    ThemeConfig  `json:"theme"`
	Input    InputConfig  `json:"input"`
	System   SystemConfig `json:"system"`
}

// --- Renderer ---------------------------------------------------------------

type RenderConfig struct {
	FontName   string `json:"font_name"`
	CellWidth  int    `json:"cell_width"`
	CellHeight int    `json:"cell_height"`
}

// --- Theme ------------------------------------------------------------------

type ThemeConfig struct {
	Name            string        `json:"name"`
	Palette         [256][3]uint8 `json:"palette"`
	CursorShape     string        `json:"cursor_shape"`
	CursorBlink     bool          `json:"cursor_blink"`
	CursorBlinkRate int           `json:"cursor_blink_rate"`
	CursorColor     [3]uint8      `json:"cursor_color"`
	CursorOpacity   uint8         `json:"cursor_opacity"`
}

// --- Input ------------------------------------------------------------------

type InputConfig struct {
	CtrlShiftCopy bool `json:"ctrl_shift_copy"`
	MouseEnabled  bool `json:"mouse_enabled"`
}

// --- System -----------------------------------------------------------------

type SystemConfig struct {
	DefaultShell string `json:"default_shell"`
	ScrollStep   int    `json:"scroll_step"` // number of lines per scroll event
}

// --- Default Configuration --------------------------------------------------

func DefaultConfig() *RootConfig {
	return &RootConfig{
		Version: "1.0",
		Renderer: RenderConfig{
			FontName:   "basic",
			CellWidth:  7,
			CellHeight: 14,
		},
		Theme: ThemeConfig{
			Name:            "default",
			CursorShape:     "block",
			CursorBlink:     false,
			CursorBlinkRate: 500,
			CursorColor:     [3]uint8{200, 200, 200},
			CursorOpacity:   200,
		},
		Input: InputConfig{
			CtrlShiftCopy: true,
			MouseEnabled:  true,
		},
		System: SystemConfig{
			DefaultShell: "/bin/bash",
			ScrollStep:   5,
		},
	}
}

// --- Accessors --------------------------------------------------------------

// Accessor methods for configuration values
func (r *RootConfig) GetCursorBlinkRate() int {
        return r.Theme.CursorBlinkRate
}

func (r *RootConfig) GetScrollStep() int {
        if r.System.ScrollStep <= 0 {
                return 5
        }
        return r.System.ScrollStep
}

func (r *RootConfig) GetDefaultShell() string {
        if r.System.DefaultShell == "" {
                return "/bin/bash"
        }
        return r.System.DefaultShell
}

