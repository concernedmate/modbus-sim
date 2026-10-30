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
	"go.bug.st/serial"
)

const (
	CONN_TYPE_TCP    = "TCP"
	CONN_TYPE_SERIAL = "SERIAL"
)

var CONN_TYPES = []string{
	CONN_TYPE_TCP,
	CONN_TYPE_SERIAL,
}
var PROTOCOL_TYPES = []string{
	modbus.TCP,
	modbus.RTU,
}
var FUNCTION_CODES_STR = []string{
	"READ_COIL",
	"READ_DISCRETE_INPUT",
	"READ_HOLDING_REGISTERS",
	"READ_INPUT_REGISTERS",
}
var FUNCTION_CODES_INT = []uint8{
	modbus.READ_COIL,
	modbus.READ_DISCRETE_INPUT,
	modbus.READ_HOLDING_REGISTERS,
	modbus.READ_INPUT_REGISTERS,
}

type WindowModscan struct {
	title string

	function_code uint8

	connection string // TCP or SERIAL

	host_addr      string   // tcp host:port
	available_port []string // serial device
	selected_port  int32
	baud_rate      int32

	slave_id int32
	register int32
	quantity int32
	protocol string

	rows []*g.TableRowWidget

	cancel  context.CancelFunc
	log_msg []string
}

func (w *WindowModscan) callback(result map[int][]byte, err error) {
	if err != nil {
		w.log("%v", err)
		return
	}

	rows := make([]*g.TableRowWidget, 0, w.quantity+1)
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

	w.rows = rows
}
func (w *WindowModscan) start() {
	if w.slave_id < 0 || w.slave_id > math.MaxUint8 {
		w.log("invalid slave_id value (must be between 0-255)")
		return
	}
	if w.quantity < 0 || w.quantity > math.MaxUint16 {
		w.log("invalid quantity value (must be between 0-65535)")
		return
	}
	if _, err := modbus.ParseRegisterAddr(int(w.register)); err != nil {
		w.log("invalid start_register value: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	if w.rows == nil {
		w.rows = make([]*g.TableRowWidget, 0, w.quantity)
	} else {
		w.rows = w.rows[:0]
	}

	if w.connection == CONN_TYPE_TCP {
		go func() {
			w.log("connection started")

			if err := modbus.ConnectEthernet(
				ctx, modbus.ConnectEthernetInput{
					Host:         w.host_addr,
					Protocol:     w.protocol,
					FunctionCode: w.function_code,
					SlaveID:      uint8(w.slave_id),
					Register:     int(w.register),
					Qty:          uint16(w.quantity),
					Callback:     w.callback,
				},
			); err != nil {
				w.log("failed to connect: %v", err)
				w.stop()
			}
			w.log("connection closed")
		}()
	} else {
		go func() {
			if int(w.selected_port) >= len(w.available_port) {
				w.log("no serial device detected")
				w.stop()
				return
			}

			w.log("connection started")
			if err := modbus.ConnectSerial(
				ctx, modbus.ConnectSerialInput{
					PortName:     w.available_port[w.selected_port],
					BaudRate:     int(w.baud_rate),
					Protocol:     w.protocol,
					FunctionCode: w.function_code,
					SlaveID:      uint8(w.slave_id),
					Register:     int(w.register),
					Qty:          uint16(w.quantity),
					Callback:     w.callback,
				},
			); err != nil {
				w.log("failed to connect: %v", err)
				w.stop()
			}
			w.log("connection closed")
		}()
	}
}
func (w *WindowModscan) stop() {
	if w.cancel == nil {
		return
	}
	w.cancel()
	w.cancel = nil
}
func (w *WindowModscan) log(format string, a ...any) {
	msg := fmt.Sprintf("[%s] %s", time.Now().Format(time.TimeOnly), fmt.Sprintf(format, a...))
	slices.Reverse(w.log_msg)
	if len(w.log_msg) < 100 {
		w.log_msg = append(w.log_msg, msg)
	} else {
		copy(w.log_msg, w.log_msg[1:])
		w.log_msg[len(w.log_msg)-1] = msg
	}
	slices.Reverse(w.log_msg)
}

func (w *WindowModscan) Title() string {
	return w.title
}
func (window *WindowModscan) Layout() []g.Widget {
	return []g.Widget{
		// config
		g.Style().SetDisabled(window.cancel != nil).To(
			g.Custom(func() {
				var rows = make([]*g.TableRowWidget, 0, 10)

				switch window.connection {
				case CONN_TYPE_SERIAL:
					var preview = ""
					if len(window.available_port) > int(window.selected_port) {
						preview = window.available_port[window.selected_port]
					}
					rows = append(rows,
						g.TableRow(g.Label("Device"), g.Combo("USB", preview, window.available_port, &window.selected_port)),
						g.TableRow(g.Label("Baud Rate"), g.InputInt(&window.baud_rate)),
					)
				case CONN_TYPE_TCP:
					rows = append(rows,
						g.TableRow(g.Label("Modbus Host"), g.InputText(&window.host_addr)),
						g.TableRow(g.Dummy(0, 0), g.Style().SetDisabled(true).To(g.InputText(new(string)))),
					)
				}

				rows = append(rows,
					g.TableRow(g.Label("Slave ID"), g.InputInt(&window.slave_id)),
					g.TableRow(g.Label("Start Register"), g.InputInt(&window.register)),
					g.TableRow(g.Label("Read Quantity"), g.InputInt(&window.quantity)),
					g.TableRow(g.Label("Function Code"),
						g.Custom(func() {
							var selected_fc_type int32 = 0
							var preview = ""
							if i := slices.Index(FUNCTION_CODES_INT, window.function_code); i != -1 {
								preview = FUNCTION_CODES_STR[i]
							}
							g.Combo("", preview, FUNCTION_CODES_STR, &selected_fc_type).OnChange(func() {
								window.function_code = FUNCTION_CODES_INT[selected_fc_type]
							}).Build()
						}),
					),
					g.TableRow(g.Label("Protocol"),
						g.Style().SetDisabled(window.connection == CONN_TYPE_SERIAL).To(
							g.Custom(func() {
								var selected_protocol int32 = 0
								g.Combo("", window.protocol, PROTOCOL_TYPES, &selected_protocol).OnChange(func() {
									if window.connection == CONN_TYPE_SERIAL {
										window.protocol = modbus.RTU
									} else {
										window.protocol = PROTOCOL_TYPES[selected_protocol]
									}
								}).Build()
							}),
						),
					),
					g.TableRow(g.Label("Connection"),
						g.Custom(func() {
							var selected_conn_type int32 = 0
							g.Combo("", window.connection, CONN_TYPES, &selected_conn_type).OnChange(func() {
								switch CONN_TYPES[selected_conn_type] {
								case CONN_TYPE_SERIAL:
									ports, err := serial.GetPortsList()
									if err != nil {
										window.log("failed to get available com devices: %v", err)
										return
									}
									window.available_port = ports
									window.connection = CONN_TYPE_SERIAL
									window.protocol = modbus.RTU
								case CONN_TYPE_TCP:
									window.connection = CONN_TYPE_TCP
								}
							}).Build()
						}),
					),
				)

				g.Table().Rows(rows...).Flags(g.TableFlagsNoClip | g.TableFlagsBorders).Build()
			}),
		),

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
			g.TabBar().TabItems(g.TabItemf("SlaveID: %d", window.slave_id).Layout(g.Table().Size(g.Auto, h/2).Rows(window.rows...))).Build()
		}),

		g.Separator(),
		// history log
		g.Label("LOG:"),
		g.ListBox(window.log_msg),
	}
}

func CreateWindowModscan(title string) WindowModscan {
	return WindowModscan{
		title:         title,
		function_code: modbus.READ_HOLDING_REGISTERS,

		connection:     CONN_TYPE_TCP,
		host_addr:      "127.0.0.1:3000",
		available_port: nil,
		selected_port:  0,
		baud_rate:      9600,

		protocol: modbus.TCP,

		slave_id: 1,
		register: 400001,
		quantity: 3,
		log_msg:  make([]string, 0, 100),

		rows: make([]*g.TableRowWidget, 0, 100),
	}
}
