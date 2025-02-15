package widgets

import (
	"errors"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/gocan"
	"github.com/roffe/gocan/adapter"
	"github.com/roffe/txlogger/pkg/layout"
	"go.bug.st/serial/enumerator"
)

const (
	prefsAdapter = "adapter"
	prefsPort    = "port"
	prefsSpeed   = "speed"
	prefsDebug   = "debug"
)

var portSpeeds = []string{"9600", "19200", "38400", "57600", "115200", "230400", "460800", "921600", "1mbit", "2mbit", "3mbit"}

type CanSettingsWidget struct {
	widget.BaseWidget
	app             fyne.App
	container       *fyne.Container
	adapterSelector *widget.Select
	debugCheckbox   *widget.Check
	portSelector    *widget.Select
	speedSelector   *widget.Select
	refreshBtn      *widget.Button
}

func NewCanSettingsWidget(app fyne.App) *CanSettingsWidget {
	csw := &CanSettingsWidget{
		app: app,
	}
	csw.ExtendBaseWidget(csw)
	csw.adapterSelector = widget.NewSelect(adapter.List(), func(s string) {
		if info, found := adapter.GetAdapterMap()[s]; found {
			app.Preferences().SetString(prefsAdapter, s)
			if info.RequiresSerialPort {
				csw.portSelector.Enable()
				csw.speedSelector.Enable()
				return
			}
			csw.portSelector.Disable()
			csw.speedSelector.Disable()
		}
	})

	csw.portSelector = widget.NewSelect(csw.listPorts(), func(s string) {
		app.Preferences().SetString(prefsPort, s)
	})
	csw.speedSelector = widget.NewSelect(portSpeeds, func(s string) {
		app.Preferences().SetString(prefsSpeed, s)
	})

	csw.debugCheckbox = widget.NewCheck("Debug", func(b bool) {
		app.Preferences().SetBool(prefsDebug, b)
	})

	csw.refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		csw.portSelector.Options = csw.listPorts()
		csw.portSelector.Refresh()
	})

	csw.container = container.NewVBox(
		container.NewBorder(
			nil,
			nil,
			layout.NewFixedWidth(75, widget.NewLabel("Adapter")),
			csw.debugCheckbox,
			csw.adapterSelector,
		),
		container.NewBorder(
			nil,
			nil,
			layout.NewFixedWidth(75, widget.NewLabel("Port")),
			csw.refreshBtn,
			csw.portSelector,
		),
		container.NewBorder(
			nil,
			nil,
			layout.NewFixedWidth(75, widget.NewLabel("Speed")),
			nil,
			csw.speedSelector,
		),
	)

	csw.loadPrefs()
	return csw
}

func (c *CanSettingsWidget) Disable() {
	c.adapterSelector.Disable()
	c.portSelector.Disable()
	c.speedSelector.Disable()
	c.debugCheckbox.Disable()
	c.refreshBtn.Disable()
}

func (c *CanSettingsWidget) Enable() {
	c.adapterSelector.Enable()
	c.portSelector.Enable()
	c.speedSelector.Enable()
	c.debugCheckbox.Enable()
	c.refreshBtn.Enable()

	if info, found := adapter.GetAdapterMap()[c.adapterSelector.Selected]; found {
		if info.RequiresSerialPort {
			c.portSelector.Enable()
			c.speedSelector.Enable()
		} else {
			c.portSelector.Disable()
			c.speedSelector.Disable()
		}
	}
}

func (cs *CanSettingsWidget) loadPrefs() {
	if adapter := cs.app.Preferences().String(prefsAdapter); adapter != "" {
		cs.adapterSelector.SetSelected(adapter)
	}
	if port := cs.app.Preferences().String(prefsPort); port != "" {
		cs.portSelector.SetSelected(port)
	}
	if speed := cs.app.Preferences().String(prefsSpeed); speed != "" {
		cs.speedSelector.SetSelected(speed)
	}
	if debug := cs.app.Preferences().Bool(prefsDebug); debug {
		cs.debugCheckbox.SetChecked(debug)
	}
}

func (cs *CanSettingsWidget) GetAdapter(ecuType string, logger func(string)) (gocan.Adapter, error) {
	baudstring := cs.speedSelector.Selected
	switch baudstring {
	case "1mbit":
		baudstring = "1000000"
	case "2mbit":
		baudstring = "2000000"
	case "3mbit":
		baudstring = "3000000"
	}

	baudrate, err := strconv.Atoi(baudstring)

	if cs.adapterSelector.Selected == "" {
		return nil, errors.New("No adapter selected") //lint:ignore ST1005 This is ok
	}

	if adapter.GetAdapterMap()[cs.adapterSelector.Selected].RequiresSerialPort {
		if cs.portSelector.Selected == "" {
			return nil, errors.New("No port selected") //lint:ignore ST1005 This is ok

		}
		if cs.speedSelector.Selected == "" {
			return nil, errors.New("No speed selected") //lint:ignore ST1005 This is ok
		}
	}

	if err != nil {
		if cs.speedSelector.Selected != "" {
			return nil, err
		}
	}

	var canFilter []uint32

	switch ecuType {
	case "T7":
		if strings.HasPrefix(cs.adapterSelector.Selected, "STN") || strings.HasPrefix(cs.adapterSelector.Selected, "OBDLink") {
			canFilter = []uint32{0x238, 0x258, 0x270}
		} else {
			canFilter = []uint32{0x1A0, 0x238, 0x258, 0x270, 0x280, 0x3A0}
		}
	case "T8":
		canFilter = []uint32{0x7e8}
	}

	return adapter.New(
		cs.adapterSelector.Selected,
		&gocan.AdapterConfig{
			Port:         cs.portSelector.Selected,
			PortBaudrate: baudrate,
			CANRate:      500,
			CANFilter:    canFilter,
			OnMessage:    logger,
			Debug:        cs.debugCheckbox.Checked,
			OnError: func(err error) {
				logger(err.Error())
			},
		},
	)
}

func (cs *CanSettingsWidget) listPorts() []string {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		//m.output(err.Error())
		return []string{}
	}
	if len(ports) == 0 {
		//m.output("No serial ports found!")
		return []string{}
	}
	for _, port := range ports {
		//m.output(fmt.Sprintf("Found port: %s", port.Name))
		if port.IsUSB {
			//m.output(fmt.Sprintf("  USB ID     %s:%s", port.VID, port.PID))
			//m.output(fmt.Sprintf("  USB serial %s", port.SerialNumber))
			portsList = append(portsList, port.Name)
		}
	}
	return portsList
}

func (cs *CanSettingsWidget) CreateRenderer() fyne.WidgetRenderer {
	return &CanSettingsWidgetRenderer{
		cs,
	}
}

type CanSettingsWidgetRenderer struct {
	*CanSettingsWidget
}

func (cs *CanSettingsWidgetRenderer) Layout(size fyne.Size) {
	cs.container.Resize(size)
}

func (cs *CanSettingsWidgetRenderer) MinSize() fyne.Size {
	return cs.container.MinSize()
}

func (cs *CanSettingsWidgetRenderer) Refresh() {
}

func (cs *CanSettingsWidgetRenderer) Destroy() {
}

func (cs *CanSettingsWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{cs.container}
}
