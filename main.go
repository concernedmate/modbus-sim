package main

import (
	"modbus-sim/window"

	g "github.com/AllenDang/giu"
)

type AppState struct {
	windows        []window.Window
	is_about_shown bool
}

func (data *AppState) toggle_about() {
	data.is_about_shown = !data.is_about_shown
}

var APP_STATE AppState

func init() {
	APP_STATE.windows = append(APP_STATE.windows, new(window.CreateWindowModscan("scanner")))
	APP_STATE.windows = append(APP_STATE.windows, new(window.CreateWindowModsim("simulator")))
}

func loop() {
	g.SingleWindowWithMenuBar().Layout(
		// menu bar
		g.MenuBar().Layout(
			g.MenuItem("About").OnClick(APP_STATE.toggle_about),
		),

		g.Custom(func() {
			tabs := make([]*g.TabItemWidget, 0, len(APP_STATE.windows))
			for _, window := range APP_STATE.windows {
				tabs = append(tabs, g.TabItem(window.Title()).Layout(window.Layout()...))
			}
			g.TabBar().TabItems(tabs...).Build()
		}),
	)
	if APP_STATE.is_about_shown {
		w, h := g.GetAvailableRegion()
		g.Window("Modbus Simulator v0.0.1").Pos(w/2, h/2).Flags(
			g.WindowFlagsNoDocking | g.WindowFlagsNoResize | g.WindowFlagsNoCollapse,
		).Layout(
			g.Align(g.AlignCenter).To(
				g.Label("Giu (github.com/AllenDang/giu) by (Allen Dang)"),
				g.Label("gopcua (github.com/gopcua/opcua) by (The gopcua authors)"),
				g.Dummy(g.Auto, 32),
				g.Label("made by concernedmate (github.com/concernedmate)"),
				g.Button("CLOSE").OnClick(APP_STATE.toggle_about),
			),
		)
	}
}
func main() {
	w := g.NewMasterWindow("Modbus", 1280, 720, g.MasterWindowFlagsTransparent)
	w.Run(loop)
}
