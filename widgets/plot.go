// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"fmt"
	"image"
	"math"
	"time"

	. "github.com/reaalkhalil/termui"
)

// Plot has two modes: line(default) and scatter.
// Plot also has two marker types: braille(default) and dot.
// A single braille character is a 2x4 grid of dots, so using braille
// gives 2x X resolution and 4x Y resolution over dot mode.
type Plot struct {
	Block

	Data       [][]float64
	DataLabels []string
	MaxVal     float64
	MinVal     float64

	LineColors []Color
	AxesColor  Color // TODO
	ShowAxes   bool

	Marker          PlotMarker
	DotMarkerRune   rune
	PlotType        PlotType
	HorizontalScale int
	DrawDirection   DrawDirection // TODO
}

const (
	xAxisLabelsHeight = 1
	yAxisLabelsWidth  = 4
	xAxisLabelsGap    = 2
	yAxisLabelsGap    = 1
)

type PlotType uint

const (
	LineChart PlotType = iota
	ScatterPlot
	CandleStickPlot
	LineChartScaled
	ScatterPlotScaled
)

type PlotMarker uint

const (
	MarkerBraille PlotMarker = iota
	MarkerDot
)

type DrawDirection uint

const (
	DrawLeft DrawDirection = iota
	DrawRight
)

func NewPlot() *Plot {
	return &Plot{
		Block:           *NewBlock(),
		LineColors:      Theme.Plot.Lines,
		AxesColor:       Theme.Plot.Axes,
		Marker:          MarkerBraille,
		DotMarkerRune:   DOT,
		Data:            [][]float64{},
		HorizontalScale: 1,
		DrawDirection:   DrawRight,
		ShowAxes:        true,
		PlotType:        LineChart,
	}
}

func (self *Plot) renderBraille(buf *Buffer, drawArea image.Rectangle, minVal, maxVal float64) {
	canvas := NewCanvas()
	canvas.Rectangle = drawArea

	switch self.PlotType {
	case ScatterPlot:
		for i, line := range self.Data {
			for j, val := range line {
				height := int((val / maxVal) * float64(drawArea.Dy()-1))
				canvas.SetPoint(
					image.Pt(
						(drawArea.Min.X+(j*self.HorizontalScale))*2,
						(drawArea.Max.Y-height-1)*4,
					),
					SelectColor(self.LineColors, i),
				)
			}
		}
	case ScatterPlotScaled:
		for i, line := range self.Data {
			for j, val := range line {
				height := int(((val - minVal) / maxVal) * float64(drawArea.Dy()-1))
				canvas.SetPoint(
					image.Pt(
						(drawArea.Min.X+(j*self.HorizontalScale))*2,
						(drawArea.Max.Y-height-1)*4,
					),
					SelectColor(self.LineColors, i),
				)
			}
		}
	case LineChart:
		for i, line := range self.Data {
			previousHeight := int((line[1] / maxVal) * float64(drawArea.Dy()-1))
			for j, val := range line[1:] {
				height := int((val / maxVal) * float64(drawArea.Dy()-1))
				canvas.SetLine(
					image.Pt(
						(drawArea.Min.X+(j*self.HorizontalScale))*2,
						(drawArea.Max.Y-previousHeight-1)*4,
					),
					image.Pt(
						(drawArea.Min.X+((j+1)*self.HorizontalScale))*2,
						(drawArea.Max.Y-height-1)*4,
					),
					SelectColor(self.LineColors, i),
				)
				previousHeight = height
			}
		}
	case LineChartScaled:
		for i, line := range self.Data {
			previousHeight := int((line[1] - minVal) / (maxVal - minVal) * float64(drawArea.Dy()-1))
			for j, val := range line[1:] {
				height := int((val - minVal) / (maxVal - minVal) * float64(drawArea.Dy()-1))
				canvas.SetLine(
					image.Pt(
						(drawArea.Min.X+(j*self.HorizontalScale))*2,
						(drawArea.Max.Y-previousHeight-1)*4,
					),
					image.Pt(
						(drawArea.Min.X+((j+1)*self.HorizontalScale))*2,
						(drawArea.Max.Y-height-1)*4,
					),
					SelectColor(self.LineColors, i),
				)
				previousHeight = height
			}
		}
	}

	canvas.Draw(buf)
}

const (
	CSStick            = '│'
	CSCandle           = '┃'
	CSHalfTop          = '╽'
	CSHalfBottom       = '╿'
	CSHalfCandleTop    = '╻'
	CSHalfCandleBottom = '╹'
	CSHalfStickTop     = '╷'
	CSHalfStickBottom  = '╵'
	CSNothing          = ' '
)

type Candle struct {
	Time   time.Time `json:"time"`
	Low    float64   `json:"low"`
	High   float64   `json:"high"`
	Open   float64   `json:"open"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

func (self *Plot) renderDot(buf *Buffer, drawArea image.Rectangle, minVal, maxVal float64) {
	switch self.PlotType {
	case CandleStickPlot:
		var cc []Candle
		for i, d := range self.Data {
			if len(cc) == 0 {
				cc = make([]Candle, len(d))
			}
			for j, n := range d {
				switch i {
				case 0:
					cc[j].Open = n
				case 1:
					cc[j].High = n
				case 2:
					cc[j].Low = n
				case 3:
					cc[j].Close = n
				}
			}
		}

		for j, c := range cc {
			llH := ((c.Low - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1)
			uuH := ((c.High - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1)
			lH := ((math.Min(c.Open, c.Close) - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1)
			uH := ((math.Max(c.Open, c.Close) - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1)

			for cy := drawArea.Min.Y - 1; cy < drawArea.Max.Y; cy++ {
				color := ColorRed
				if c.Close >= c.Open {
					color = ColorGreen
				}

				ch := renderCandleAt(llH, uuH, lH, uH, drawArea.Max.Y-1-cy)
				if ch == CSNothing {
					color = ColorWhite
				}

				point := image.Pt(drawArea.Min.X+(j*self.HorizontalScale), cy)
				if point.In(drawArea) {
					buf.SetCell(
						NewCell(ch, NewStyle(color)),
						point,
					)
				}
			}
		}

	case ScatterPlot:
		for i, line := range self.Data {
			for j, val := range line {
				height := int((val / maxVal) * float64(drawArea.Dy()-1))
				point := image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height)
				if point.In(drawArea) {
					buf.SetCell(
						NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, i))),
						point,
					)
				}
			}
		}
	case ScatterPlotScaled:
		for i, line := range self.Data {
			for j, val := range line {
				height := int(((val - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1))
				point := image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height)
				if point.In(drawArea) {
					buf.SetCell(
						NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, i))),
						point,
					)
				}
			}
		}
	case LineChart:
		for i, line := range self.Data {
			for j := 0; j < len(line) && j*self.HorizontalScale < drawArea.Dx(); j++ {
				val := line[j]
				height := int((val / maxVal) * float64(drawArea.Dy()-1))
				buf.SetCell(
					NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, i))),
					image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height),
				)
			}
		}
	case LineChartScaled:
		for i, line := range self.Data {
			for j := 0; j < len(line) && j*self.HorizontalScale < drawArea.Dx(); j++ {
				val := line[j]
				height := int(((val - minVal) / (maxVal - minVal)) * float64(drawArea.Dy()-1))
				buf.SetCell(
					NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, i))),
					image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height),
				)
			}
		}
	}
}

func renderCandleAt(llH, uuH, lH, uH float64, heightUnit int) rune {
	heightUnit64 := float64(heightUnit)

	scaledTopStick := uuH
	scaledTopCandle := uH
	scaledBottomStick := llH
	scaledBottomCandle := lH

	if math.Ceil(scaledTopStick) >= heightUnit64 && heightUnit64 >= math.Floor(scaledTopCandle) {
		if scaledTopCandle-heightUnit64 > 0.75 {
			return CSCandle
		} else if (scaledTopCandle - heightUnit64) > 0.25 {
			if (scaledTopStick - heightUnit64) > 0.75 {
				return CSHalfTop
			}
			return CSHalfCandleTop
		} else {
			if (scaledTopStick - heightUnit64) > 0.75 {
				return CSStick
			} else if (scaledTopStick - heightUnit64) > 0.25 {
				return CSHalfStickTop
			} else {
				return CSNothing
			}
		}
	} else if math.Floor(scaledTopCandle) >= heightUnit64 && heightUnit64 >= math.Ceil(scaledBottomCandle) {
		return CSCandle
	} else if math.Ceil(scaledBottomCandle) >= heightUnit64 && heightUnit64 >= math.Floor(scaledBottomStick) {
		if (scaledBottomCandle - heightUnit64) < 0.25 {
			return CSCandle
		} else if (scaledBottomCandle - heightUnit64) < 0.75 {
			if (scaledBottomStick - heightUnit64) < 0.25 {
				return CSHalfBottom
			}
			return CSHalfCandleBottom
		} else {
			if (scaledBottomStick - heightUnit64) < 0.25 {
				return CSStick
			} else if (scaledBottomStick - heightUnit64) < 0.75 {
				return CSHalfStickBottom
			} else {
				return CSNothing
			}
		}
	}
	return CSNothing
}

func (self *Plot) plotAxes(buf *Buffer, minVal, maxVal float64) {
	// draw origin cell
	buf.SetCell(
		NewCell(BOTTOM_LEFT, NewStyle(ColorWhite)),
		image.Pt(self.Inner.Min.X+yAxisLabelsWidth, self.Inner.Max.Y-xAxisLabelsHeight-1),
	)
	// draw x axis line
	for i := yAxisLabelsWidth + 1; i < self.Inner.Dx(); i++ {
		buf.SetCell(
			NewCell(HORIZONTAL_DASH, NewStyle(ColorWhite)),
			image.Pt(i+self.Inner.Min.X, self.Inner.Max.Y-xAxisLabelsHeight-1),
		)
	}
	// draw y axis line
	for i := 0; i < self.Inner.Dy()-xAxisLabelsHeight-1; i++ {
		buf.SetCell(
			NewCell(VERTICAL_DASH, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Min.X+yAxisLabelsWidth, i+self.Inner.Min.Y),
		)
	}
	// draw x axis labels
	// draw 0
	buf.SetString(
		"0",
		NewStyle(ColorWhite),
		image.Pt(self.Inner.Min.X+yAxisLabelsWidth, self.Inner.Max.Y-1),
	)
	// draw rest
	for x := self.Inner.Min.X + yAxisLabelsWidth + (xAxisLabelsGap)*self.HorizontalScale + 1; x < self.Inner.Max.X-1; {
		label := fmt.Sprintf(
			"%d",
			(x-(self.Inner.Min.X+yAxisLabelsWidth)-1)/(self.HorizontalScale)+1,
		)
		buf.SetString(
			label,
			NewStyle(ColorWhite),
			image.Pt(x, self.Inner.Max.Y-1),
		)
		x += (len(label) + xAxisLabelsGap) * self.HorizontalScale
	}
	// draw y axis labels
	// TODO:   check self.PlotType to either use minVal or not
	verticalScale := (maxVal - minVal) / float64(self.Inner.Dy()-xAxisLabelsHeight-1)
	for i := 0; i*(yAxisLabelsGap+1) < self.Inner.Dy()-1; i++ {
		buf.SetString(
			fmt.Sprintf("%.2f", minVal+float64(i)*verticalScale*(yAxisLabelsGap+1)),
			NewStyle(ColorWhite),
			image.Pt(self.Inner.Min.X, self.Inner.Max.Y-(i*(yAxisLabelsGap+1))-2),
		)
	}
}

func (self *Plot) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	maxVal := self.MaxVal
	minVal := self.MinVal
	if maxVal == 0 {
		maxVal, _ = GetMaxFloat64From2dSlice(self.Data)
	}
	if minVal == 0 {
		minVal, _ = GetMinFloat64From2dSlice(self.Data)
	}

	if self.ShowAxes {
		self.plotAxes(buf, minVal, maxVal)
	}

	drawArea := self.Inner
	if self.ShowAxes {
		drawArea = image.Rect(
			self.Inner.Min.X+yAxisLabelsWidth+1, self.Inner.Min.Y,
			self.Inner.Max.X, self.Inner.Max.Y-xAxisLabelsHeight-1,
		)
	}

	switch self.Marker {
	case MarkerBraille:
		self.renderBraille(buf, drawArea, minVal, maxVal)
	case MarkerDot:
		self.renderDot(buf, drawArea, minVal, maxVal)
	}
}
