// Copyright 2021 The oksvg Authors. All rights reserved.
//
// clip.go implements parsing and rasterization of SVG clipPath elements.

package oksvg

import (
	"encoding/xml"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/srwiley/rasterx"
)

// clipPathF parses a clipPath start element. Its child elements are collected
// into the SvgClip by readStartElement while IconCursor.clip is set.
var clipPathF svgFunc = func(c *IconCursor, attrs []xml.Attr) error {
	clip := &SvgClip{}
	for _, attr := range attrs {
		if attr.Name.Local == "clipPathUnits" &&
			strings.TrimSpace(attr.Value) == "objectBoundingBox" {
			clip.UnitsObjectBoundingBox = true
		}
	}
	c.clip = clip
	for _, attr := range attrs {
		if attr.Name.Local == "id" && attr.Value != "" {
			c.icon.Clips[attr.Value] = clip
		}
	}
	return nil
}

// appendClipPath stores a path parsed inside a clipPath element into the
// current clip definition. It returns true when a clipPath is being parsed,
// whether or not a path was collected.
func (c *IconCursor) appendClipPath() bool {
	if c.clip == nil {
		return false
	}
	if len(c.Path) > 0 {
		pathCopy := make(rasterx.Path, len(c.Path))
		copy(pathCopy, c.Path)
		c.clip.Paths = append(c.clip.Paths,
			SvgPath{PathStyle: c.StyleStack[len(c.StyleStack)-1], Path: pathCopy})
		c.Path = c.Path[:0]
	}
	return true
}

// readClipURL extracts the referenced id from an SVG url(#id) value. The
// boolean result is false for clip-path="none" or any value that is not a
// fragment URL reference.
func readClipURL(v string) (id string, ok bool) {
	v = strings.TrimSpace(v)
	if v == "none" || !strings.HasPrefix(v, "url(") || !strings.HasSuffix(v, ")") {
		return "", false
	}
	urlStr := strings.TrimSpace(v[4 : len(v)-1])
	urlStr = strings.Trim(urlStr, "'\"")
	if !strings.HasPrefix(urlStr, "#") {
		return "", false
	}
	return urlStr[1:], true
}

// resolveClipRefs resolves the clip-path id references collected while
// parsing against the clipPath definitions of the icon. clip-path="none" and
// references to missing or empty clip definitions are dropped.
func (s *SvgIcon) resolveClipRefs() {
	for i := range s.SVGPaths {
		refs := s.SVGPaths[i].clipRefs
		if len(refs) == 0 {
			continue
		}
		clips := make([]*SvgClip, 0, len(refs))
		for _, ref := range refs {
			clip, ok := s.Clips[ref]
			if !ok || clip == nil {
				continue // missing reference: no clipping
			}
			if len(clip.Paths) == 0 {
				// An empty clip path means nothing is painted.
				s.SVGPaths[i].clipEmpty = true
				continue
			}
			clips = append(clips, clip)
		}
		s.SVGPaths[i].clipPaths = clips
	}
}

// drawTransformedClip renders svgp clipped by svgp.clipPaths. The clipping
// regions are rasterized into alpha masks and intersected; the path itself is
// first drawn onto a layer the size of the destination and then composited
// through the combined mask. An empty clip region suppresses drawing entirely.
func (svgp *SvgPath) drawTransformedClip(r *rasterx.Dasher, opacity float64, t rasterx.Matrix2D) {
	gv, ok := r.Scanner.(*rasterx.ScannerGV)
	if !ok {
		// Only the ScannerGV rasterizer exposes the destination image
		// needed for mask based clipping; fall back to unclipped drawing
		// for other Scanner implementations.
		svgp.drawTransformedNoClip(r, opacity, t)
		return
	}
	dst := gv.Dest
	dstBounds := dst.Bounds()
	w, h := dstBounds.Dx(), dstBounds.Dy()
	// All off-screen buffers use a (0,0)-origin rectangle so their Pix
	// indexing matches pixel coordinates one to one.
	bufBounds := image.Rect(0, 0, w, h)

	var mask *image.Alpha
	for ci, clip := range svgp.clipPaths {
		clipMask, empty := svgp.rasterizeClip(clip, t, bufBounds)
		if empty {
			return // empty clip path: the whole element must not be painted
		}
		if ci == 0 {
			// The first mask is the initial region; further masks are
			// intersected with it by taking the per-pixel alpha minimum.
			mask = clipMask
			continue
		}
		intersectAlpha(mask, clipMask)
	}

	layer := image.NewRGBA(bufBounds)
	layerScanner := rasterx.NewScannerGV(w, h, layer, bufBounds)
	layerRaster := rasterx.NewDasher(w, h, layerScanner)
	// Draw the element unclipped onto the layer; masking is applied by the
	// DrawMask call below.
	svgp.drawTransformedNoClip(layerRaster, opacity, t)
	// Align the (0,0)-origin buffers with the destination rectangle.
	draw.DrawMask(dst, dstBounds, layer, image.Point{}, mask, image.Point{}, draw.Over)
}

// rasterizeClip rasterizes one clip definition into an alpha image. It
// returns an empty flag when the clip region does not cover any pixel.
func (svgp *SvgPath) rasterizeClip(clip *SvgClip,
	t rasterx.Matrix2D, bounds image.Rectangle) (*image.Alpha, bool) {
	mask := image.NewAlpha(bounds)
	scanner := rasterx.NewScannerGV(bounds.Dx(), bounds.Dy(), mask, bounds)
	filler := rasterx.NewFiller(bounds.Dx(), bounds.Dy(), scanner)
	filler.SetColor(color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})

	// For objectBoundingBox units the clip path coordinates are fractions of
	// the element bounding box. userSpaceBounds returns the element bounds in
	// the same pixel space the mask is rasterized in (i.e. after t is
	// applied), so scaling and translating the unit square onto those bounds
	// already yields pixel-space clip geometry and t must not be applied a
	// second time.
	clipToMask := t
	if clip.UnitsObjectBoundingBox {
		ext := svgp.userSpaceBounds(t, bounds.Dx(), bounds.Dy())
		minX, minY := float64(ext.Min.X)/64.0, float64(ext.Min.Y)/64.0
		maxX, maxY := float64(ext.Max.X)/64.0, float64(ext.Max.Y)/64.0
		clipToMask = rasterx.Matrix2D{
			A: maxX - minX, C: 0, E: minX,
			B: 0, D: maxY - minY, F: minY,
		}
	}

	adder := rasterx.MatrixAdder{}
	for i := range clip.Paths {
		cp := &clip.Paths[i]
		// userSpaceOnUse clip geometry shares the element user space and is
		// transformed by t; objectBoundingBox geometry is mapped directly to
		// the pixel-space element bounds. The clip content keeps its own
		// element transform in both cases.
		adder.M = clipToMask.Mult(cp.mAdder.M)
		adder.Adder = filler
		cp.Path.AddTo(&adder)
	}
	filler.Draw()

	empty := true
	for _, a := range mask.Pix {
		if a != 0 {
			empty = false
			break
		}
	}
	return mask, empty
}

// intersectAlpha stores the per-pixel minimum of a and b back into a,
// implementing the intersection of two alpha masks.
func intersectAlpha(a, b *image.Alpha) {
	pa, pb := a.Pix, b.Pix
	for i := range pa {
		if pb[i] < pa[i] {
			pa[i] = pb[i]
		}
	}
}
