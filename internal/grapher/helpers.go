package grapher

import (
	"fmt"
)

func RenderTable(name string, duration StartupDuration, keys []StartupTag) string {
	body := fmt.Sprintf(`
		<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
		<TR><TD>Step</TD><TD>%s</TD></TR>
		<TR><TD>Dur</TD><TD>%v</TD></TR>`, name, duration)

	for _, t := range keys {
		// for k, v := range t {
		tagLine := fmt.Sprintf("<TR><TD>TagKey: %s</TD><TD>TagValue: %s</TD></TR>", t.Key, t.Value)
		body = body + tagLine
		// }
	}

	body = body + "</TABLE>"

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
