package files

import (
	"strconv"
	"strings"
)

type pathComponent struct {
	componentType string
	text          string
}

func (p *pathComponent) String() string {
	switch p.componentType {
	case "obj":
		return "." + p.text
	case "arr":
		return "[" + p.text + "]"
	default:
		return ""
	}
}

type pathTracker struct {
	components []pathComponent
}

func (p *pathTracker) String() string {
	parts := make([]string, 0, len(p.components)+1)

	parts = append(parts, "$")

	for _, c := range p.components {
		parts = append(parts, c.String())
	}

	return "YAML path \"" + strings.Join(parts, "") + "\""
}

func (p *pathTracker) PushObj(fieldName string) *pathTracker {
	pt := &pathTracker{
		components: make([]pathComponent, len(p.components)+1),
	}

	copy(pt.components, p.components)
	pt.components[len(p.components)] = pathComponent{
		componentType: "obj",
		text:          fieldName,
	}

	return pt
}

func (p *pathTracker) PushArr(index int) *pathTracker {
	pt := &pathTracker{
		components: make([]pathComponent, len(p.components)+1),
	}

	copy(pt.components, p.components)
	pt.components[len(p.components)] = pathComponent{
		componentType: "arr",
		text:          strconv.FormatInt(int64(index), 10),
	}

	return pt
}
