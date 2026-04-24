package cmdutil

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// PickerItem represents one option in an interactive picker.
type PickerItem struct {
	Value string // the value returned on selection (e.g. "42")
	Label string // displayed label (e.g. "#42  Fix login bug (alice → main)")
}

// RunPicker shows an interactive filterable selection list.
// Returns the selected item's Value, or error if cancelled.
func RunPicker(title string, items []PickerItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select from")
	}

	opts := make([]huh.Option[string], len(items))
	for i, item := range items {
		opts[i] = huh.NewOption(item.Label, item.Value)
	}

	var selected string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(title).
			Options(opts...).
			Value(&selected).
			Filtering(true),
	))

	if err := form.Run(); err != nil {
		return "", err
	}
	return selected, nil
}
