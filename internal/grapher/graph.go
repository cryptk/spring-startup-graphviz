package grapher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/lucasb-eyer/go-colorful"
)

type Grapher struct {
	gviz           *graphviz.Graphviz
	graph          *cgraph.Graph
	data           StartupResponse
	filterDuration time.Duration
	minDuration    time.Duration
	maxDuration    time.Duration
}

func New(filterDuraton time.Duration) (*Grapher, error) {
	gviz := graphviz.New()
	graph, err := gviz.Graph()
	if err != nil {
		return nil, err
	}

	grapher := &Grapher{
		gviz:           gviz,
		graph:          graph,
		filterDuration: filterDuraton,
	}

	return grapher, nil
}

func (g *Grapher) ParseURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	startupResponse, err := g.ParseText(body)
	if err != nil {
		return err
	}

	g.data = startupResponse

	return nil
}

func (g *Grapher) ParseText(jsonData []byte) (StartupResponse, error) {
	var startupResponse StartupResponse
	if err := json.Unmarshal(jsonData, &startupResponse); err != nil {
		return StartupResponse{}, err
	}

	return startupResponse, nil
}

func (g *Grapher) Generate() error {

	// First we assemble an array of all events
	nodes := make([]graphNode, len(g.data.Timeline.Events))
	for _, event := range g.data.Timeline.Events {
		nodes[event.StartupStep.ID] = graphNode{nil, event, false, nil}
		if event.Duration.Duration < g.minDuration {
			g.minDuration = event.Duration.Duration
		}
		if event.Duration.Duration > g.maxDuration {
			g.maxDuration = event.Duration.Duration
		}
	}

	// Next we wire up the parent references in our array
	// We do this in a separate pass because we need to ensure the parent node already exists in the array
	// so that we can get a valid reference to the parent.
	// There is no guarantee that the events are "in order" in the source data
	for _, node := range nodes {
		if node.event.StartupStep.ParentID != 0 {
			nodes[node.event.StartupStep.ID].parent = &nodes[node.event.StartupStep.ParentID]
		}
	}

	// Then we mark any nodes that should be rendered in the graph
	//
	// If an event is longer than our filterDuration it should be rendered
	// If an event has a (grand)child that would be rendered, it should be rendered
	//
	// This ensures that we will render nodes that have children that would be rendered even if the parent would not otherwise be rendered.
	for _, node := range nodes {
		if node.event.Duration.Duration > g.filterDuration {
			markParentsRecursive(&node)
		}
	}

	// Next we add Graphviz nodes for every event we determined should be rendered
	// TODO: This could possibly be handled at the same time as we mark the node for rendering
	for _, node := range nodes {
		if node.shouldRender {
			n, err := g.graph.CreateNode(fmt.Sprint(node.event.StartupStep.ID))
			if err != nil {
				return err
			}
			n.SetShape(cgraph.BoxShape)
			n.SetStyle(cgraph.FilledNodeStyle)
			bgColor, err := g.getBGColor(node.event.Duration.Duration)
			if err != nil {
				return err
			}
			n.SetFillColor(bgColor)
			n.SetLabel(g.graph.StrdupHTML(RenderTable(node.event.StartupStep.Name, node.event.Duration.Duration, node.event.StartupStep.Tags)))
			if err != nil {
				return err
			}
			nodes[node.event.StartupStep.ID].node = n
		}
	}

	// Now that we know that all of the required nodes are in the graph
	// we can create the edges between them
	for _, node := range nodes {
		if node.shouldRender && node.parent != nil {
			_, err := g.graph.CreateEdge("test", node.parent.node, node.node)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Grapher) RenderDOT() (bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := g.gviz.Render(g.graph, "dot", &buf); err != nil {
		return bytes.Buffer{}, err
	}
	return buf, nil
}

func (g *Grapher) RenderSVGFile(filepath string) error {
	if err := g.gviz.RenderFilename(g.graph, graphviz.SVG, filepath); err != nil {
		return err
	}
	return nil
}

func (g *Grapher) RenderPNGFile(filepath string) error {
	if err := g.gviz.RenderFilename(g.graph, graphviz.PNG, filepath); err != nil {
		return err
	}
	return nil
}

func (g *Grapher) getBGColor(duration time.Duration) (string, error) {
	// TODO: allow overriding this color value with a config param
	colorFast, err := colorful.Hex("#00FF00")
	if err != nil {
		return "", err
	}

	// TODO: allow overriding this color value with a config param
	colorSlow, err := colorful.Hex("#FF0000")
	if err != nil {
		return "", err
	}

	// TODO: The 0.25 scaling value is just a random value that makes the colors look "good".
	// The startup events typically have a single event for spring.context.refresh that lasts for
	// almost the entire startup duration.  Ideally, we would determine what our "max full red" value is
	// more intelligently... perhaps using the 90th percentile value or something.
	// we should also have a config param to allow overriding whatever logic we decide to use
	fadePercentage := math.Min(duration.Seconds()/(g.maxDuration.Seconds()*0.25), 1.0)
	return colorFast.BlendLab(colorSlow, fadePercentage).Clamped().Hex(), nil
}

func (g *Grapher) Close() error {
	if err := g.graph.Close(); err != nil {
		return err
	}
	if err := g.gviz.Close(); err != nil {
		return err
	}
	return nil
}
