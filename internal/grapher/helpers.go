package grapher

import (
	"fmt"
	"html"
	"time"
)

func RenderTable(name string, duration time.Duration, keys []StartupTag) string {
	name = html.EscapeString(name)
	body := fmt.Sprintf(`
		<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD>Step</TD><TD>%s</TD></TR>
		<TR><TD>Dur</TD><TD>%v</TD></TR>`, name, duration)

	for _, t := range keys {
		key := html.EscapeString(t.Key)
		value := html.EscapeString(t.Value)
		tagLine := fmt.Sprintf("\n<TR><TD>TagKey: %s</TD><TD>TagValue: %s</TD></TR>", key, value)
		body = body + tagLine
	}

	body = body + "\n</TABLE>"

	return body
}

func markParentsRecursive(node *graphNode) {
	if !node.shouldRender {
		// log.Printf("Marking node for rendering: %d: %s", node.event.StartupStep.ID, node.event.Duration)
		node.shouldRender = true
	}
	if node.parent != nil {
		markParentsRecursive(node.parent)
	}
}
