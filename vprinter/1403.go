package vprinter

// Copyright 2021 Matthew R. Wilson <mwilson@mattwilson.org>
//
// This file is part of virtual1403
// <https://github.com/racingmars/virtual1403>.
//
// virtual1403 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// virtual1403 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with virtual1403. If not, see <https://www.gnu.org/licenses/>.

import (
	"io"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

const maxLinesPerPage = 66
const maxLineCharacters = 132

// Colors
const (
	// greenDarkR = 70
	// greenDarkG = 150
	// greenDarkB = 70

	greenDarkR = 99
	greenDarkG = 182
	greenDarkB = 99

	// greenLightR = 195
	// greenLightG = 229
	// greenLightB = 195

	greenLightR = 219
	greenLightG = 240
	greenLightB = 219
)

// our implementation of the Job interface simulating an IBM 1403 printer.
type virtual1403 struct {
	pdf        *gofpdf.Fpdf
	font       []byte
	curLine    int
	leftMargin float64
}

// Page size
const (
	v1403W = 1071 // 14 7/8 inches wide
	v1403H = 792  // 11 inches high
)

const v1430FontSize = 11.4

func New1403(font []byte) (Job, error) {
	j := &virtual1403{
		font: font,
	}

	j.pdf = gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: v1403W, Ht: v1403H},
	})

	j.pdf.SetMargins(0, 0, 0)
	j.pdf.SetAutoPageBreak(false, 0)

	// Despite the documentation, it appears that AddUTF8Font takes the font
	// directly, not the JSON file generated by makefont. We also, then, have
	// to assume the font just magically gets embedded automatically.
	j.pdf.AddUTF8FontFromBytes("userfont", "", j.font)

	// We will dynamically determine how wide 132 characters of the chosen
	// font is so that we can correctly position (center) the output area on
	// the page. The left margin of our text output area will be the center
	// of the page minus half of the line width.
	j.pdf.SetFont("userfont", "", v1430FontSize)
	j.leftMargin = v1403W/2 - determineLineWidth(j.pdf)/2

	j.pdf.SetHeaderFunc(func() { drawBackground(j.pdf) })

	j.NewPage()

	return j, nil
}

func (job *virtual1403) AddLine(s string) {
	if job.curLine >= maxLinesPerPage {
		job.NewPage()
	}
	if len(s) > maxLineCharacters {
		s = s[0:maxLineCharacters]
	}
	// 1403 only had capital letters
	s = strings.ToUpper(s)
	job.pdf.SetXY(job.leftMargin, float64(job.curLine*12)+.25)
	job.pdf.CellFormat(0, 12, s, "", 0, "LM", false, 0, "")
	job.curLine++
}

func (job *virtual1403) NewPage() {
	job.pdf.AddPage()
	job.pdf.SetFont("userfont", "", v1430FontSize)
	// simulating a 1403 with form control that skips the first 5 physically
	// printable lines.
	job.curLine = 5
}

func (job *virtual1403) EndJob(w io.Writer) error {
	return job.pdf.Output(w)
}

func drawBackground(pdf *gofpdf.Fpdf) {
	const feedHoleRadius = 5.5

	// Alignment fiducial
	pdf.SetDrawColor(greenDarkR, greenDarkG, greenDarkB)
	pdf.SetLineWidth(.7)
	pdf.Line(20, 54-feedHoleRadius*2, 20, 54+feedHoleRadius*2)
	pdf.Line(20-feedHoleRadius*2, 54, 20+feedHoleRadius*2, 54)
	pdf.SetLineWidth(1.5)
	pdf.Circle(20, 54, feedHoleRadius+.6, "D")

	// Draw tractor feed circles -- top and bottom holes are larger
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetFillColor(230, 230, 230)
	pdf.SetLineWidth(.75)
	// Top holes
	y := float64(18 + 18*2*0)
	pdf.Circle(20, y, feedHoleRadius+1, "FD")
	pdf.Circle(v1403W-20, y, feedHoleRadius+1, "FD")
	// Bottom holes
	y = float64(18 + 18*2*21)
	pdf.Circle(20, y, feedHoleRadius+1, "FD")
	pdf.Circle(v1403W-20, y, feedHoleRadius+1, "FD")
	for i := 1; i < 21; i++ {
		y := float64(18 + 18*2*i)
		pdf.Circle(20, y, feedHoleRadius, "FD")
		pdf.Circle(v1403W-20, y, feedHoleRadius, "FD")
	}

	// Draw form number - 1412THE
	pdf.SetTextColor(greenDarkR, greenDarkG, greenDarkB)
	pdf.SetFont("helvetica", "", 7)
	pdf.SetXY(v1403W-4, 55)
	pdf.TransformBegin()
	pdf.TransformRotate(-90, v1403W-4, 55)
	pdf.CellFormat(0, 7, "1412THE", "", 0, "", false, 0, "")
	pdf.TransformEnd()

	// Print area alignment arrows
	pdf.SetFillColor(greenLightR, greenLightG, greenLightB)
	// Left side
	pdf.Polygon([]gofpdf.PointType{
		{X: 40 + 2, Y: 72 - 11},
		{X: 40 + 2 + 5, Y: 72},
		{X: 40 + 2 + 5*2, Y: 72 - 11},
	}, "F")
	// Right side
	pdf.Polygon([]gofpdf.PointType{
		{X: v1403W - 40 - 2, Y: 72 - 11},
		{X: v1403W - 40 - 2 - 5, Y: 72},
		{X: v1403W - 40 - 2 - 5*2, Y: 72 - 11},
	}, "F")

	// There is an outline "1" above the bottom-right tractor feed hole.
	// Drawing it will be a manual exercise. I designed the 1 on graph paper,
	// so all the numbers in the following path drawing is based on my
	// translation of the graph paper grid to the PDF coordinates.
	const bX float64 = v1403W - 20 // bottom-left of "1"
	const bY float64 = v1403H - 29 // bottom-left of "1"
	const bU float64 = 0.6         // 1 grid unit in points
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(greenDarkR, greenDarkG, greenDarkB)
	pdf.MoveTo(bX+bU*5, bY-bU*17)
	pdf.LineTo(bX+bU*5, bY-bU*3.5)
	pdf.LineTo(bX, bY-bU*3.5)
	pdf.LineTo(bX, bY)
	pdf.LineTo(bX+bU*14, bY)
	pdf.LineTo(bX+bU*14, bY-bU*3.5)
	pdf.LineTo(bX+bU*9, bY-bU*3.5)
	pdf.LineTo(bX+bU*9, bY-bU*24)
	pdf.LineTo(bX+bU*8, bY-bU*24)
	pdf.CurveTo(bX+bU*6, bY-bU*20.5, bX, bY-bU*19) // top curved segment
	pdf.LineTo(bX, bY-bU*15)
	pdf.CurveTo(bX+bU*3.5, bY-bU*15.5, bX+bU*5, bY-bU*17) // bottom curved segment
	pdf.ClosePath()
	pdf.DrawPath("D")

	// Green bars. We are drawing the fill separate from the lines, because it
	// looks like the horizontal lines are slightly heavier than the vertical
	// lines.
	pdf.SetFillColor(greenLightR, greenLightG, greenLightB)
	for i := 0; i < 10; i++ {
		pdf.Rect(40, float64(72+i*72)-.5, v1403W-80, 36, "F")
	}

	// Horizontal lines. The top line and bottom line are full width to cap
	// the margin number columns, the other lines are only as wide as the
	// greenbars. The extra 0.25-point wiggle-room is to make the corners of
	// the vertical and horizontal lines square with each other.
	pdf.SetDrawColor(greenDarkR, greenDarkG, greenDarkB)
	pdf.SetLineWidth(.7)
	pdf.Line(30-.25, 72-.5, v1403W-30+.25, 72-.5)             // top
	pdf.Line(30-.25, v1403H-1-.5, v1403W-30+.25, v1403H-1-.5) // bottom
	for i := 0; i < 20; i++ {
		pdf.Line(40, float64(72+36*i)-.5, v1403W-40, float64(72+36*i)-.5)
	}

	// Vertical lines
	pdf.SetDrawColor(greenDarkR, greenDarkG, greenDarkB)
	pdf.SetLineWidth(.5)
	pdf.Line(30, 72-.5, 30, v1403H-1-.5)
	pdf.Line(40, 72-.5, 40, v1403H-1-.5)

	pdf.Line(v1403W-30, 72-.5, v1403W-30, v1403H-1-.5)
	pdf.Line(v1403W-40, 72-.5, v1403W-40, v1403H-1-.5)

	// Left margin numbers
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(greenDarkR, greenDarkG, greenDarkB)
	for i := 0; i < 60; i++ {
		pdf.SetXY(30, float64(72+i*12))
		// The centering of the margin numbers looks better if we use
		// *slightly* different width for the cell for single- versus double-
		// digit numbers.
		w := 9.7
		if i < 9 {
			w = 10
		}
		pdf.CellFormat(w, 12, strconv.Itoa(i+1), "", 0, "CM", false, 0, "")
	}

	// Right margin numbers
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(greenDarkR, greenDarkG, greenDarkB)
	for i := 0; i < 80; i++ {
		pdf.SetXY(v1403W-40, float64(72+i*9))
		// The centering of the margin numbers looks better if we use
		// *slightly* different width for the cell for single- versus double-
		// digit numbers.
		w := 9.7
		if i < 9 {
			w = 10
		}
		pdf.CellFormat(w, 9, strconv.Itoa(i+1), "", 0, "CM", false, 0, "")
	}

	pdf.SetTextColor(0, 0, 0)
}

func determineLineWidth(pdf *gofpdf.Fpdf) float64 {
	const linechars = 132
	var dummyline [linechars]byte
	for i := 0; i < linechars; i++ {
		dummyline[i] = ' '
	}
	return pdf.GetStringWidth(string(dummyline[:]))
}
