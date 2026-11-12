// Copyright 2021 The oksvg Authors. All rights reserved.
//
// clip_test.go verifies SVG clipPath rendering: userSpaceOnUse and
// objectBoundingBox units, inheritance through <g>, intersection of nested
// clips, clip-path="none", missing references and empty clip paths.

package oksvg_test

import (
	"image"
	"strings"
	"testing"

	. "github.com/srwiley/oksvg"
	. "github.com/srwiley/rasterx"
)

// render100 reads svg with ReadIconStream and renders it into a 100x100
// RGBA image at 1:1 scale.
func render100(t *testing.T, svg string) *image.RGBA {
	t.Helper()
	icon, err := ReadIconStream(strings.NewReader(svg), WarnErrorMode)
	if err != nil {
		t.Fatalf("ReadIconStream failed: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	scanner := NewScannerGV(100, 100, img, img.Bounds())
	raster := NewDasher(100, 100, scanner)
	icon.Draw(raster, 1.0)
	return img
}

const userSpaceCircleClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="circleClip">
      <circle cx="50" cy="50" r="20"/>
    </clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#circleClip)"/>
</svg>`

const objectBoundingBoxClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="obbClip" clipPathUnits="objectBoundingBox">
      <circle cx="0.5" cy="0.5" r="0.25"/>
    </clipPath>
  </defs>
  <rect x="10" y="10" width="80" height="80" fill="black" clip-path="url(#obbClip)"/>
</svg>`

const nestedClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="leftHalf"><rect x="0" y="0" width="50" height="100"/></clipPath>
    <clipPath id="topHalf"><rect x="0" y="0" width="100" height="50"/></clipPath>
  </defs>
  <g clip-path="url(#leftHalf)">
    <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#topHalf)"/>
  </g>
</svg>`

const gInheritedClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="circleClip">
      <circle cx="50" cy="50" r="20"/>
    </clipPath>
  </defs>
  <g clip-path="url(#circleClip)">
    <rect x="0" y="0" width="100" height="100" fill="black"/>
  </g>
</svg>`

const clipPathNone = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="circleClip">
      <circle cx="50" cy="50" r="20"/>
    </clipPath>
  </defs>
  <g clip-path="url(#circleClip)">
    <rect x="0" y="0" width="100" height="100" fill="black" clip-path="none"/>
  </g>
</svg>`

const missingClipRef = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#doesNotExist)"/>
</svg>`

const emptyClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="emptyClip"></clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#emptyClip)"/>
</svg>`

const strokeClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c"><circle cx="50" cy="50" r="20"/></clipPath>
  </defs>
  <rect x="5" y="5" width="90" height="90" fill="none" stroke="black" stroke-width="4" clip-path="url(#c)"/>
</svg>`

const styleAttrClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c"><circle cx="50" cy="50" r="20"/></clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" style="clip-path:url(#c)"/>
</svg>`

const quotedURLClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c"><circle cx="50" cy="50" r="20"/></clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path='url("#c")'/>
</svg>`

const multiShapeClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c">
      <circle cx="30" cy="50" r="15"/>
      <circle cx="70" cy="50" r="15"/>
    </clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#c)"/>
</svg>`

const transformedClipContent = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c">
      <g transform="translate(20,20)"><rect x="0" y="0" width="40" height="40"/></g>
    </clipPath>
  </defs>
  <rect x="0" y="0" width="100" height="100" fill="black" clip-path="url(#c)"/>
</svg>`

const useWithClip = `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100">
  <defs>
    <clipPath id="c"><circle cx="50" cy="50" r="20"/></clipPath>
    <rect id="r" x="0" y="0" width="100" height="100" fill="black" clip-path="url(#c)"/>
  </defs>
  <use href="#r"/>
</svg>`

// scaledUserSpaceClip uses a 50x50 viewBox scaled 2x via SetTarget, so the
// radius-10 user-space circle becomes a radius-20 pixel circle.
const scaledUserSpaceClip = `<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" viewBox="0 0 50 50">
  <defs>
    <clipPath id="c"><circle cx="25" cy="25" r="10"/></clipPath>
  </defs>
  <rect x="0" y="0" width="50" height="50" fill="black" clip-path="url(#c)"/>
</svg>`

const scaledOBBClip = `<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50" viewBox="0 0 50 50">
  <defs>
    <clipPath id="c" clipPathUnits="objectBoundingBox">
      <circle cx="0.5" cy="0.5" r="0.5"/>
    </clipPath>
  </defs>
  <rect x="0" y="0" width="50" height="50" fill="black" clip-path="url(#c)"/>
</svg>`

func alphaAt(img *image.RGBA, x, y int) uint8 {
	return img.RGBAAt(x, y).A
}

func assertTransparent(t *testing.T, img *image.RGBA, pts ...image.Point) {
	t.Helper()
	for _, p := range pts {
		if a := alphaAt(img, p.X, p.Y); a != 0 {
			t.Errorf("pixel (%d,%d): want alpha 0, got %d", p.X, p.Y, a)
		}
	}
}

func assertOpaque(t *testing.T, img *image.RGBA, pts ...image.Point) {
	t.Helper()
	for _, p := range pts {
		if a := alphaAt(img, p.X, p.Y); a != 255 {
			t.Errorf("pixel (%d,%d): want alpha 255, got %d", p.X, p.Y, a)
		}
	}
}

// TestUserSpaceOnUseClip renders a 100x100 rect clipped by a radius-20
// circle: the corners and everything outside the circle must stay fully
// transparent while the circle interior is fully opaque.
func TestUserSpaceOnUseClip(t *testing.T) {
	img := render100(t, userSpaceCircleClip)

	// Four corners must be untouched.
	assertTransparent(t, img,
		image.Pt(0, 0), image.Pt(99, 0), image.Pt(0, 99), image.Pt(99, 99))
	// Points well outside the r=20 circle (center 50,50).
	assertTransparent(t, img,
		image.Pt(5, 5), image.Pt(50, 5), image.Pt(95, 50), image.Pt(50, 95))
	// Circle interior must be fully painted.
	assertOpaque(t, img,
		image.Pt(50, 50), image.Pt(55, 50), image.Pt(45, 45), image.Pt(60, 55))
}

// TestObjectBoundingBoxClip checks clipPathUnits="objectBoundingBox": the
// r=0.25 circle expressed in the 80x80 rect bounding box maps to a radius-20
// circle centered at (50,50) in pixel space.
func TestObjectBoundingBoxClip(t *testing.T) {
	img := render100(t, objectBoundingBoxClip)

	assertTransparent(t, img,
		image.Pt(0, 0), image.Pt(99, 99), image.Pt(10, 10), image.Pt(50, 5),
		image.Pt(90, 90), image.Pt(5, 50))
	assertOpaque(t, img,
		image.Pt(50, 50), image.Pt(55, 50), image.Pt(45, 45))
}

// TestNestedClipIntersection verifies that a clip inherited from a <g>
// parent and the element's own clip are intersected: only the top-left
// quadrant may be painted.
func TestNestedClipIntersection(t *testing.T) {
	img := render100(t, nestedClip)

	assertOpaque(t, img,
		image.Pt(5, 5), image.Pt(25, 25), image.Pt(49, 49))
	// Right half (outside the inherited left-half clip).
	assertTransparent(t, img,
		image.Pt(75, 25), image.Pt(95, 5), image.Pt(99, 49))
	// Bottom half (outside the element's top-half clip).
	assertTransparent(t, img,
		image.Pt(25, 75), image.Pt(5, 95), image.Pt(49, 99))
	// Bottom-right quadrant is outside both.
	assertTransparent(t, img, image.Pt(75, 75), image.Pt(99, 99))
}

// TestGroupClipInheritance verifies that clip-path on a <g> applies to its
// descendants.
func TestGroupClipInheritance(t *testing.T) {
	img := render100(t, gInheritedClip)

	assertTransparent(t, img,
		image.Pt(0, 0), image.Pt(99, 99), image.Pt(5, 5), image.Pt(50, 95))
	assertOpaque(t, img, image.Pt(50, 50), image.Pt(55, 50), image.Pt(45, 45))
}

// TestClipPathNone verifies clip-path="none" disables clipping, including a
// clip inherited from an ancestor group.
func TestClipPathNone(t *testing.T) {
	img := render100(t, clipPathNone)

	assertOpaque(t, img,
		image.Pt(0, 0), image.Pt(99, 0), image.Pt(0, 99), image.Pt(99, 99),
		image.Pt(50, 50))
}

// TestMissingClipReference verifies that a reference to a missing clipPath
// does not clip the element.
func TestMissingClipReference(t *testing.T) {
	img := render100(t, missingClipRef)

	assertOpaque(t, img,
		image.Pt(0, 0), image.Pt(99, 99), image.Pt(50, 50))
}

// TestEmptyClipPath verifies that an empty clip path suppresses painting of
// the whole element.
func TestEmptyClipPath(t *testing.T) {
	img := render100(t, emptyClip)

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if a := alphaAt(img, x, y); a != 0 {
				t.Fatalf("empty clip path painted pixel (%d,%d) with alpha %d", x, y, a)
			}
		}
	}
}

// TestStrokeClip verifies clipping applies to stroked (fill="none") geometry
// as well as filled geometry.
func TestStrokeClip(t *testing.T) {
	img := render100(t, strokeClip)

	assertTransparent(t, img, image.Pt(0, 0), image.Pt(99, 99))
	// The rect border is well outside the radius-20 clip circle.
	assertTransparent(t, img, image.Pt(8, 8), image.Pt(92, 92))
}

// TestClipPathInStyleAttribute verifies clip-path given inside a style="..."
// attribute works the same as the presentation attribute.
func TestClipPathInStyleAttribute(t *testing.T) {
	img := render100(t, styleAttrClip)

	assertTransparent(t, img, image.Pt(0, 0), image.Pt(99, 99), image.Pt(50, 5))
	assertOpaque(t, img, image.Pt(50, 50))
}

// TestQuotedClipURL verifies url("#id") with quotes is parsed.
func TestQuotedClipURL(t *testing.T) {
	img := render100(t, quotedURLClip)

	assertTransparent(t, img, image.Pt(0, 0), image.Pt(99, 99))
	assertOpaque(t, img, image.Pt(50, 50))
}

// TestMultiShapeClipPath verifies that all shapes inside one clipPath form
// the union of the clip region.
func TestMultiShapeClipPath(t *testing.T) {
	img := render100(t, multiShapeClip)

	assertOpaque(t, img, image.Pt(30, 50), image.Pt(70, 50))
	assertTransparent(t, img,
		image.Pt(5, 5), image.Pt(50, 50), image.Pt(99, 99))
}

// TestTransformedClipContent verifies transforms on the clipPath contents
// (here a translated <g>) are honored.
func TestTransformedClipContent(t *testing.T) {
	img := render100(t, transformedClipContent)

	assertOpaque(t, img, image.Pt(40, 40))
	assertTransparent(t, img, image.Pt(10, 10), image.Pt(80, 80))
}

// TestClipPathWithUse verifies a clip-path on an element referenced through
// <use> is applied.
func TestClipPathWithUse(t *testing.T) {
	img := render100(t, useWithClip)

	assertTransparent(t, img, image.Pt(0, 0), image.Pt(99, 99))
	assertOpaque(t, img, image.Pt(50, 50))
}

// renderScaled reads svg and renders it into a 100x100 image with the icon
// scaled 2x through SetTarget.
func renderScaled(t *testing.T, svg string) *image.RGBA {
	t.Helper()
	icon, err := ReadIconStream(strings.NewReader(svg), WarnErrorMode)
	if err != nil {
		t.Fatalf("ReadIconStream failed: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	scanner := NewScannerGV(100, 100, img, img.Bounds())
	raster := NewDasher(100, 100, scanner)
	icon.SetTarget(0, 0, 100, 100)
	icon.Draw(raster, 1.0)
	return img
}

// TestScaledUserSpaceClip verifies userSpaceOnUse clip geometry scales with
// the icon transform from SetTarget.
func TestScaledUserSpaceClip(t *testing.T) {
	img := renderScaled(t, scaledUserSpaceClip)

	assertOpaque(t, img, image.Pt(50, 50))
	assertTransparent(t, img, image.Pt(0, 0), image.Pt(75, 50), image.Pt(99, 99))
}

// TestScaledObjectBoundingBoxClip verifies objectBoundingBox clip geometry
// scales with the element bounds under SetTarget.
func TestScaledObjectBoundingBoxClip(t *testing.T) {
	img := renderScaled(t, scaledOBBClip)

	assertOpaque(t, img, image.Pt(50, 50))
	assertTransparent(t, img, image.Pt(0, 0), image.Pt(99, 99))
}

// TestClipNonZeroOrigin renders into an image whose bounds do not start at
// (0,0), verifying off-screen buffers are aligned with the destination.
func TestClipNonZeroOrigin(t *testing.T) {
	icon, err := ReadIconStream(strings.NewReader(userSpaceCircleClip), WarnErrorMode)
	if err != nil {
		t.Fatalf("ReadIconStream failed: %v", err)
	}
	bounds := image.Rect(10, 20, 110, 120)
	img := image.NewRGBA(bounds)
	scanner := NewScannerGV(100, 100, img, bounds)
	raster := NewDasher(100, 100, scanner)
	icon.Draw(raster, 1.0)

	// Circle center (50,50) in icon space maps to (60,70) in the image.
	if a := alphaAt(img, 60, 70); a != 255 {
		t.Errorf("circle center alpha=%d want 255", a)
	}
	// Corner (0,0) in icon space maps to (10,20).
	if a := alphaAt(img, 10, 20); a != 0 {
		t.Errorf("corner alpha=%d want 0", a)
	}
}
