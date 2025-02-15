package widgets

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type VBar struct {
	face        *canvas.Rectangle
	bar         *canvas.Rectangle
	titleText   *canvas.Text
	displayText *canvas.Text
	bars        []*canvas.Line

	cfg *VBarConfig

	value float64

	canvas fyne.CanvasObject
}

type VBarConfig struct {
	Title    string
	Min, Max float64
	Steps    int
	Minsize  fyne.Size
}

func NewVBar(cfg *VBarConfig) *VBar {
	s := &VBar{
		cfg: cfg,
	}
	if s.cfg.Steps == 0 {
		s.cfg.Steps = 10
	}
	s.canvas = s.render()
	return s
}

func (s *VBar) render() *fyne.Container {
	s.face = &canvas.Rectangle{StrokeColor: color.RGBA{0x80, 0x80, 0x80, 0x80}, FillColor: color.RGBA{0x00, 0x00, 0x00, 0x00}, StrokeWidth: 2}
	s.bar = &canvas.Rectangle{StrokeColor: color.RGBA{0x2C, 0xA5, 0x00, 0x80}, FillColor: color.RGBA{0x2C, 0xA5, 0x00, 0x80}}

	s.titleText = &canvas.Text{Text: s.cfg.Title, Color: color.RGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF}, TextSize: 25}
	s.titleText.TextStyle.Monospace = true
	s.titleText.Alignment = fyne.TextAlignCenter

	s.displayText = &canvas.Text{Text: "0", Color: color.RGBA{R: 0xF0, G: 0xF0, B: 0xF0, A: 0xFF}, TextSize: 25}
	s.displayText.TextStyle.Monospace = true
	s.displayText.Alignment = fyne.TextAlignCenter

	bar := container.NewWithoutLayout(s.face)
	for i := int(s.cfg.Steps + 1); i > 0; i-- {
		line := &canvas.Line{StrokeColor: color.RGBA{byte(i * 10), 0xE5 - byte(i*10), 0x00, 0xFF}, StrokeWidth: 2}
		s.bars = append(s.bars, line)
		bar.Add(line)
	}
	bar.Objects = append(bar.Objects, s.bar, s.titleText, s.displayText)
	bar.Layout = s
	return bar
}

func (s *VBar) Layout(_ []fyne.CanvasObject, space fyne.Size) {
	diameter := space.Width
	middle := diameter / 2
	heightFactor := float32(space.Height) / float32(s.cfg.Steps)

	s.face.Resize(space)

	s.titleText.Move(fyne.NewPos(middle-s.titleText.Size().Width/2, space.Height+2))

	s.displayText.Move(fyne.NewPos(space.Width/2-s.displayText.Size().Width/2, space.Height-(float32(s.value)*heightFactor)-12.5))

	s.bar.Move(fyne.NewPos(0, space.Height-float32(s.value)))

	for i, line := range s.bars {
		if i%2 == 0 {
			line.Position1 = fyne.NewPos(middle-diameter/3, float32(i)*heightFactor)
			line.Position2 = fyne.NewPos(middle+diameter/3, float32(i)*heightFactor)
			continue
		}
		line.Position1 = fyne.NewPos(middle-diameter/7, float32(i)*heightFactor)
		line.Position2 = fyne.NewPos(middle+diameter/7, float32(i)*heightFactor)
	}
	s.SetValue(s.value)
}

func (s *VBar) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return s.cfg.Minsize
}

func (s *VBar) Content() fyne.CanvasObject {
	return s.canvas
}

func (s *VBar) SetValue(value float64) {
	//if value == s.value {
	//	return
	//}
	if value > s.cfg.Max {
		value = s.cfg.Max
	}
	if value < s.cfg.Min {
		value = s.cfg.Min
	}

	s.value = value
	size := s.canvas.Size()
	heightFactor := float32(size.Height) / float32(s.cfg.Max)
	diameter := size.Width

	br := 0xA5 * (value / s.cfg.Max)
	bg := 0xA5 - br
	if bg < 0 {
		bg = 0
	}

	s.bar.FillColor = color.RGBA{byte(br), byte(bg), 0x00, 0x80}

	s.bar.Move(fyne.NewPos(diameter/8, size.Height-(float32(value)*heightFactor)))
	s.bar.Resize(fyne.NewSize(size.Width-(diameter/8*2), (float32(value) * heightFactor)))
	//s.bar.Refresh()

	s.displayText.Text = strconv.FormatFloat(value, 'f', 0, 64)
	s.displayText.Move(fyne.NewPos(size.Width/2-s.displayText.Size().Width/2, size.Height-(float32(value)*heightFactor)-12.5))
	s.displayText.Refresh()
}

func (s *VBar) Value() float64 {
	return s.value
}
