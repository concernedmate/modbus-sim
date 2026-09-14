package main

import (
	"context"
	"fmt"
	"log"
	"modbus-sim/modbus"
	"modbus-sim/window"
	"os"
	"strconv"

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
	APP_STATE.windows = append(APP_STATE.windows, new(window.CreateWindowModscan("modbus scanner")))
	APP_STATE.windows = append(APP_STATE.windows, new(window.CreateWindowModsim("modbus simulator")))
}
func args() {
	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("app modsim")
		fmt.Println("\t start modbus simulator")
		fmt.Println("app modscan ip:port slave_id addr count")
		fmt.Println("\t start modbus scanner")
		fmt.Println("\t example: app modscan 127.0.0.1:502 1 40001 10")
		os.Exit(0)
	}
	switch os.Args[1] {
	case "modsim":
		if err := modbus.ServeTCP(context.Background(), "127.0.0.1:8000", nil); err != nil {
			log.Fatal(err)
		}
	case "modscan":
		if len(os.Args) < 6 {
			fmt.Println("app modscan ip:port slave_id addr count")
			fmt.Println("\t start modbus scanner")
			fmt.Println("\t example: app modscan 127.0.0.1:502 1 40001 10")
			os.Exit(0)
		}
		addr := os.Args[2]
		slave_id, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid slave_id: %v", err))
		}
		register, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid register: %v", err))
		}
		qty, err := strconv.Atoi(os.Args[5])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid qty: %v", err))
		}
		if err := modbus.ConnectTCP(context.Background(), modbus.READ_HOLDING_REGISTERS, addr, uint8(slave_id), register, uint16(qty), nil); err != nil {
			log.Fatal(err)
		}
	}
}

func loop() {
	g.SingleWindowWithMenuBar().Layout(
		// menu bar
		g.MenuBar().Layout(
			g.MenuItem("About").OnClick(APP_STATE.toggle_about),
		),

		g.Custom(func() {
			for idx := range APP_STATE.windows {
				APP_STATE.windows[idx].Build()
			}
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
