package vprinter

// Copyright 2022 Matthew R. Wilson <mwilson@mattwilson.org>
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
	_ "embed"
	"strings"
)

var DarkGreen = ColorRGB{99, 182, 99}
var LightGreen = ColorRGB{219, 240, 219}

var DarkBlue = ColorRGB{65, 182, 255}
var LightBlue = ColorRGB{214, 239, 255}

//go:embed IBMPlexMono-Regular.ttf
var defaultFont []byte

//go:embed IBM140310Pitch-Regular-MRW.ttf
var wornFont []byte

func NewProfile(profile string, fontOverride []byte,
	sizeOverride float64) (Job, error) {

	// Some profiles use the proprietary 1403 Vintage Mono font that we can't
	// ship with the code. If the installation doesn't have that font (or
	// another font which the configuration provides), we use IBM Plex Mono
	// instead.
	tempFont := defaultFont
	if fontOverride != nil {
		tempFont = fontOverride
	}
	tempSize := 12.0
	if sizeOverride > 0 {
		tempSize = sizeOverride
	}

	switch strings.ToLower(profile) {
	case "default-green":
		return New1403(tempFont, tempSize, 6, true, true, DarkGreen, LightGreen)
	case "default-green-noskip":
		return New1403(tempFont, tempSize, 1, true, true, DarkGreen, LightGreen)
	case "default-blue":
		return New1403(tempFont, tempSize, 6, true, true, DarkBlue, LightBlue)
	case "default-blue-noskip":
		return New1403(tempFont, tempSize, 1, true, true, DarkBlue, LightBlue)
	case "default-plain":
		return New1403(tempFont, tempSize, 6, true, false, ColorRGB{}, ColorRGB{})
	case "default-plain-noskip":
		return New1403(tempFont, tempSize, 1, true, false, ColorRGB{}, ColorRGB{})
	case "retro-green":
		return New1403(wornFont, 10, 6, true, true, DarkGreen, LightGreen)
	case "retro-green-noskip":
		return New1403(wornFont, 10, 1, true, true, DarkGreen, LightGreen)
	case "retro-blue":
		return New1403(wornFont, 10, 6, true, true, DarkBlue, LightBlue)
	case "retro-blue-noskip":
		return New1403(wornFont, 10, 1, true, true, DarkBlue, LightBlue)
	case "retro-plain":
		return New1403(wornFont, 10, 6, true, false, ColorRGB{}, ColorRGB{})
	case "retro-plain-noskip":
		return New1403(wornFont, 10, 1, true, false, ColorRGB{}, ColorRGB{})
	case "modern-green":
		return New1403(defaultFont, 12.0, 6, false, true, DarkGreen, LightGreen)
	case "modern-green-skip5":
		return New1403(defaultFont, 12.0, 5, false, true, DarkGreen, LightGreen)
	case "modern-green-noskip":
		return New1403(defaultFont, 12.0, 1, false, true, DarkGreen, LightGreen)
	case "modern-blue":
		return New1403(defaultFont, 12.0, 6, false, true, DarkBlue, LightBlue)
	case "modern-blue-skip5":
		return New1403(defaultFont, 12.0, 5, false, true, DarkBlue, LightBlue)
	case "modern-blue-noskip":
		return New1403(defaultFont, 12.0, 1, false, true, DarkBlue, LightBlue)
	case "modern-plain":
		return New1403(defaultFont, 12.0, 6, false, false, ColorRGB{}, ColorRGB{})
	case "modern-plain-skip5":
		return New1403(defaultFont, 12.0, 5, false, false, ColorRGB{}, ColorRGB{})
	case "modern-plain-noskip":
		return New1403(defaultFont, 12.0, 1, false, false, ColorRGB{}, ColorRGB{})
	case "lpi8-modern-green":
		return New1403(defaultFont, 9.0, 8, false, true, DarkGreen, LightGreen)
	case "lpi8-modern-green-noskip":
		return New1403(defaultFont, 9.0, 1, false, true, DarkGreen, LightGreen)
	case "lpi8-modern-blue":
		return New1403(defaultFont, 9.0, 8, false, true, DarkBlue, LightBlue)
	case "lpi8-modern-blue-noskip":
		return New1403(defaultFont, 9.0, 1, false, true, DarkBlue, LightBlue)
	case "lpi8-modern-plain":
		return New1403(defaultFont, 9.0, 8, false, false, ColorRGB{}, ColorRGB{})
	case "lpi8-modern-plain-noskip":
		return New1403(defaultFont, 9.0, 1, false, false, ColorRGB{}, ColorRGB{})
	default:
		// default is the same as default-green
		return New1403(tempFont, tempSize, 6, true, true, DarkGreen, LightGreen)
	}
}
