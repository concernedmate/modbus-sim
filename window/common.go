package window

import g "github.com/AllenDang/giu"

type Window interface {
	Title() string
	Layout() []g.Widget
}
