package window

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"maps"
	"math"
	"modbus-sim/modbus"
	"slices"
	"strings"
	"time"

	g "github.com/AllenDang/giu"
)

type WindowModscan struct {
	title string

	modbus_function_code int32
	modbus_host_addr     string
	modbus_slave_id      int32
	modbus_register      int32
	modbus_quantity      int32
	modbus_rows          []*g.TableRowWidget

	cancel  context.CancelFunc
	log_msg []string
}

func (window *WindowModscan) callback(result map[int][]byte, err error) {
	if err != nil {
		window.log("%v", err)
		return
	}

	rows := make([]*g.TableRowWidget, 0, window.modbus_quantity+1)
	bytes_builder := strings.Builder{}

	// headers
	rows = append(rows, g.TableRow(g.Label("Address"), g.Label("Value"), g.Label("Binary")).Flags(g.TableRowFlagsHeaders))
	// values
	for _, register := range slices.Sorted(maps.Keys(result)) {
		value := result[register]

		// build representation
		bytes_builder.Reset()
		for idx, val := range hex.EncodeToString(value) {
			if idx != 0 && idx%2 == 0 {
				bytes_builder.WriteRune(' ')
			}
			bytes_builder.WriteRune(val)
		}

		// insert row
		rows = append(rows, g.TableRow(
			g.Labelf("%d", register),
			g.Labelf("%04d", binary.BigEndian.Uint16(value)),
			g.Labelf("[%s]\n", bytes_builder.String()),
		))
	}

	window.modbus_rows = rows
}
func (window *WindowModscan) start() {
	if window.modbus_slave_id < 0 || window.modbus_slave_id > math.MaxUint8 {
		window.log("invalid slave_id value (must be between 0-255)")
		return
	}
	if window.modbus_quantity < 0 || window.modbus_quantity > math.MaxUint16 {
		window.log("invalid quantity value (must be between 0-65535)")
		return
	}
	if _, err := modbus.ParseRegisterAddr(int(window.modbus_register)); err != nil {
		window.log("invalid start_register value: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	window.cancel = cancel

	go func() {
		window.log("connection started")
		if err := modbus.ConnectTCP(
			ctx, byte(window.modbus_function_code), window.modbus_host_addr, uint8(window.modbus_slave_id),
			int(window.modbus_register), uint16(window.modbus_quantity),
			window.callback,
		); err != nil {
			window.log("failed to connect: %v", err)
			window.stop()
		}
		window.log("connection closed")
	}()
}
func (window *WindowModscan) stop() {
	if window.cancel == nil {
		return
	}
	window.cancel()
	window.cancel = nil
}
func (window *WindowModscan) log(format string, a ...any) {
	msg := fmt.Sprintf("[%s] %s", time.Now().Format(time.TimeOnly), fmt.Sprintf(format, a...))
	slices.Reverse(window.log_msg)
	if len(window.log_msg) < 100 {
		window.log_msg = append(window.log_msg, msg)
	} else {
		copy(window.log_msg, window.log_msg[1:])
		window.log_msg[len(window.log_msg)-1] = msg
	}
	slices.Reverse(window.log_msg)
}
func (window *WindowModscan) Build() {
	g.Window(window.title).Size(g.GetAvailableRegion()).Layout(
		// config
		g.Custom(func() {
			g.Table().Rows(
				g.TableRow(g.Label("Modbus Host"), g.InputText(&window.modbus_host_addr)),
				g.TableRow(g.Label("Slave ID"), g.InputInt(&window.modbus_slave_id)),
				g.TableRow(g.Label("Start Register"), g.InputInt(&window.modbus_register)),
				g.TableRow(g.Label("Read Quantity"), g.InputInt(&window.modbus_quantity)),
				g.TableRow(g.Label("Function Code"), g.InputInt(&window.modbus_function_code)),
			).Flags(g.TableFlagsNoClip + g.TableFlagsBorders).Build()
		}),

		// buttons
		g.Custom(func() {
			w, _ := g.GetAvailableRegion()
			g.Row(
				g.Button("CONNECT").Size(w/2, 32).OnClick(window.start).Disabled(window.cancel != nil),
				g.Button("DISCONNECT").Size(w/2, 32).OnClick(window.stop).Disabled(window.cancel == nil),
			).Build()
		}),

		// devices
		g.Custom(func() {
			_, h := g.GetAvailableRegion()
			g.TabBar().TabItems(g.TabItemf("SlaveID: %d", window.modbus_slave_id).Layout(g.Table().Size(g.Auto, h/2).Rows(window.modbus_rows...))).Build()
		}),

		g.Separator(),
		// history log
		g.Label("LOG:"),
		g.ListBox(window.log_msg),
	)
}

func CreateWindowModscan(title string) WindowModscan {
	return WindowModscan{
		title:                title,
		modbus_function_code: modbus.READ_HOLDING_REGISTERS,
		modbus_host_addr:     "127.0.0.1:3000",
		modbus_slave_id:      1,
		modbus_register:      400001,
		modbus_quantity:      3,
		log_msg:              make([]string, 0, 100),
	}
}
