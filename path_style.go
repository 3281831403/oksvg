// Copyright 2017 The oksvg Authors. All rights reserved.
// created: 2/12/2017 by S.R.Wiley
//
// utils.go implements translation of an SVG2.0 path into a rasterx Path.

package oksvg

import (
	"image/color"

	"github.com/srwiley/rasterx"
)

// SvgClip holds the geometry of a parsed clipPath element. The paths are in
// the user coordinate system in place at the time the clipPath element is
// parsed, together with their own PathStyle transforms.
type SvgClip struct {
	// UnitsObjectBoundingBox is true when clipPathUnits="objectBoundingBox";
	// false means userSpaceOnUse (the SVG default).
	UnitsObjectBoundingBox bool
	// Paths are the renderable elements contained in the clipPath. An empty
	// Paths slice means the clip region is empty, so clipped elements are
	// not drawn at all.
	Paths []SvgPath
}

// PathStyle holds the state of the SVG style.
type PathStyle struct {
	FillOpacity, LineOpacity          float64
	LineWidth, DashOffset, MiterLimit float64
	Dash                              []float64
	UseNonZeroWinding                 bool
	fillerColor, linerColor           interface{} // either color.Color or rasterx.Gradient
	LineGap                           rasterx.GapFunc
	LeadLineCap                       rasterx.CapFunc // This is used if different than LineCap
	LineCap                           rasterx.CapFunc
	LineJoin                          rasterx.JoinMode
	mAdder                            rasterx.MatrixAdder // current transform
	// clipRefs holds the clip-path IDs that apply to this path, including
	// the ones inherited from ancestor <g> elements. An empty (but non-nil)
	// slice means the element explicitly set clip-path="none". nil means
	// no clip-path applies at all.
	clipRefs []string
	// clipPaths is populated after parsing by resolving clipRefs against
	// the SvgIcon clip definitions. An empty slice means no clipping.
	clipPaths []*SvgClip
	// clipEmpty is true when at least one resolved clip reference points at
	// an empty clipPath, in which case the element must not be painted.
	clipEmpty bool
}

// styleAttribute describes draw options, such as {"fill":"black"; "stroke":"white"}.
type styleAttribute = map[string]string

// DefaultStyle sets the default PathStyle to fill black, winding rule,
// full opacity, no stroke, ButtCap line end and Bevel line connect.
var DefaultStyle = PathStyle{1.0, 1.0, 2.0, 0.0, 4.0, nil, true,
	color.NRGBA{0x00, 0x00, 0x00, 0xff}, nil,
	nil, nil, rasterx.ButtCap, rasterx.Bevel, rasterx.MatrixAdder{M: rasterx.Identity},
	nil, nil, false}
