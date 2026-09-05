package window

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"modbus-sim/modbus"
	"strings"
	"time"

	g "github.com/AllenDang/giu"
)

type Window interface {
	Build()
}

type WindowModsim struct {
	title string
	host  string

	stop_serve context.CancelFunc

	selected_device int
	devices         []modbus.Device
	log_msg         []string
}

func (window *WindowModsim) toggle_server() {
	if window.stop_serve != nil {
		window.stop_serve()
		window.stop_serve = nil
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		window.stop_serve = cancel

		go func() {
			if err := modbus.ServeTCP(ctx, window.host, window.devices); err != nil {
				window.log("failed to start server: %v", err)
				window.toggle_server()
			}
		}()
	}

}
func (window *WindowModsim) log(format string, a ...any) {
	msg := fmt.Sprintf("[%s] %s", time.Now().Format(time.TimeOnly), fmt.Sprintf(format, a))
	if len(window.log_msg) < 100 {
		window.log_msg = append(window.log_msg, msg)
	} else {
		copy(window.log_msg, window.log_msg[1:])
		window.log_msg[len(window.log_msg)-1] = msg
	}
}
func (window *WindowModsim) Build() {
	g.Window(window.title).Layout(
		// config
		g.Custom(func() {
			g.Table().Rows(
				g.TableRow(g.Label("Modsim Host"), g.InputText(&window.host)),
				g.TableRow(g.Label("Modsim Devices"), g.Labelf("%d", len(window.devices))),
			).Flags(g.TableFlagsNoClip + g.TableFlagsBorders).Build()
		}),

		// buttons
		g.Custom(func() {
			w, _ := g.GetAvailableRegion()
			g.Row(
				g.Button("START SERVER").Size(w/2, 32).OnClick(window.toggle_server),
				g.Button("CLOSE").Size(w/2, 32),
			).Build()
		}),

		// devices
		g.Custom(func() {
			_, h := g.GetAvailableRegion()
			var items = make([]*g.TabItemWidget, 0, len(window.devices))
			var bytes = make([]byte, 2)
			var bytes_builder strings.Builder

			for idx := range len(window.devices) {
				rows := make([]*g.TableRowWidget, 0, len(window.devices[idx].HoldingRegisters))

				// headers
				rows = append(rows, g.TableRow(g.Label("Address"), g.Label("Value"), g.Label("Binary")).Flags(g.TableRowFlagsHeaders))
				// values
				for idx, val := range window.devices[idx].HoldingRegisters {
					// make binary version
					binary.BigEndian.PutUint16(bytes, val)

					// build representation
					bytes_builder.Reset()
					for idx, val := range hex.EncodeToString(bytes) {
						if idx != 0 && idx%2 == 0 {
							bytes_builder.WriteRune(' ')
						}
						bytes_builder.WriteRune(val)
					}

					// insert row
					rows = append(rows, g.TableRow(
						g.Labelf("%d", 400001+idx),
						g.Labelf("%04d", val),
						g.Labelf("[%s]\n", bytes_builder.String()),
					))
				}

				// add device
				items = append(items, g.TabItemf("SlaveID: %d", window.devices[idx].SlaveID).EventHandler(
					g.Event().OnClick(g.MouseButtonLeft, func() { window.selected_device = idx }),
				).Layout(g.Table().Size(g.Auto, h/2).Rows(rows...)))
			}

			g.TabBar().TabItems(items...).Build()
		}),

		g.Separator(),
		// history log
		g.Label("LOG:"),
		g.ListBox(window.log_msg),
	)
}

func Tes(title string) WindowModsim {
	return WindowModsim{
		title: title,
		host:  "127.0.0.1:3000",
		devices: []modbus.Device{
			modbus.NewModbusDevice(1, []uint16{1, 2, 3}, nil),
			modbus.NewModbusDevice(2, []uint16{0, 256, 512}, nil),
		},
		log_msg: make([]string, 0, 100),
	}
}
