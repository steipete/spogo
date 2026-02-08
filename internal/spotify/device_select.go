package spotify

import (
	"fmt"
	"sort"
	"strings"
)

// ResolveDeviceID resolves a user-supplied selector (name or id) to a concrete device id.
//
// Resolution order:
// - exact id match (case-insensitive)
// - exact name match (case-insensitive)
// - unique substring match on name
// - unique substring match on id
//
// If multiple devices match, returns an error describing the ambiguity.
func ResolveDeviceID(devices []Device, selector string) (string, error) {
	sel := strings.TrimSpace(selector)
	if sel == "" {
		return "", fmt.Errorf("empty device selector")
	}
	selFold := strings.ToLower(sel)

	// 1) Exact ID.
	for _, d := range devices {
		if strings.EqualFold(d.ID, sel) {
			return d.ID, nil
		}
	}
	// 2) Exact name.
	for _, d := range devices {
		if strings.EqualFold(d.Name, sel) {
			return d.ID, nil
		}
	}

	// 3) Substring name.
	nameMatches := make([]Device, 0, 4)
	for _, d := range devices {
		if d.Name == "" {
			continue
		}
		if strings.Contains(strings.ToLower(d.Name), selFold) {
			nameMatches = append(nameMatches, d)
		}
	}
	if len(nameMatches) == 1 {
		return nameMatches[0].ID, nil
	}
	if len(nameMatches) > 1 {
		return "", ambiguousDeviceSelectorError(sel, nameMatches)
	}

	// 4) Substring id.
	idMatches := make([]Device, 0, 4)
	for _, d := range devices {
		if d.ID == "" {
			continue
		}
		if strings.Contains(strings.ToLower(d.ID), selFold) {
			idMatches = append(idMatches, d)
		}
	}
	if len(idMatches) == 1 {
		return idMatches[0].ID, nil
	}
	if len(idMatches) > 1 {
		return "", ambiguousDeviceSelectorError(sel, idMatches)
	}

	return "", fmt.Errorf("device %q not found", sel)
}

func ambiguousDeviceSelectorError(selector string, matches []Device) error {
	labels := make([]string, 0, len(matches))
	for _, d := range matches {
		label := d.ID
		if strings.TrimSpace(d.Name) != "" {
			label = fmt.Sprintf("%s (%s)", d.Name, d.ID)
		}
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return fmt.Errorf("device %q is ambiguous; matches: %s", selector, strings.Join(labels, ", "))
}
