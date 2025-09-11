package theme

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var hexRe = regexp.MustCompile(`^#?([0-9a-fA-F]{6})$`)

func mustClamp01(x float64) float64 {
	if x < 0 {
		return 0
	}

	if x > 1 {
		return 1
	}

	return x
}

func hexToRGB(hex string) (r, g, b float64, err error) {
	m := hexRe.FindStringSubmatch(hex)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("invalid hex")
	}

	s := strings.ToLower(m[1])

	rv, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	gv, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	bv, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	return float64(rv) / 255.0, float64(gv) / 255.0, float64(bv) / 255.0, nil
}

func rgbToHex(r, g, b float64) string {
	r = math.Round(mustClamp01(r) * 255.0)
	g = math.Round(mustClamp01(g) * 255.0)
	b = math.Round(mustClamp01(b) * 255.0)

	return fmt.Sprintf("#%02x%02x%02x", int(r), int(g), int(b))
}

// RGB [0..1] → HSL (H in degrees [0..360), S/L in [0..1])
func rgbToHsl(r, g, b float64) (h, s, l float64) {
	maxVal := math.Max(r, math.Max(g, b))
	minVal := math.Min(r, math.Min(g, b))
	l = (maxVal + minVal) / 2

	if maxVal == minVal {
		return 0, 0, l
	}

	d := maxVal - minVal
	if l > 0.5 {
		s = d / (2 - maxVal - minVal)
	} else {
		s = d / (maxVal + minVal)
	}

	switch maxVal {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}

	h *= 60
	// normalize
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	return
}

func hslToRgb(h, s, l float64) (r, g, b float64) {
	// normalize hue defensively
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2

	var rp, gp, bp float64

	switch {
	case h < 60:
		rp, gp, bp = c, x, 0
	case h < 120:
		rp, gp, bp = x, c, 0
	case h < 180:
		rp, gp, bp = 0, c, x
	case h < 240:
		rp, gp, bp = 0, x, c
	case h < 300:
		rp, gp, bp = x, 0, c
	default:
		rp, gp, bp = c, 0, x
	}

	return rp + m, gp + m, bp + m
}

// GeneratePalette9 returns 9 swatches; index 4 is exactly the base color.
// GeneratePalette9 returns 9 hex colours where index 4 is the base.
// The lightness deltas are shaped to mimic the visual spacing in the
// reference scale (denser near the base, wider at the extremes).
func GeneratePalette9(baseHex string) ([]string, error) {
	// 1) Parse base → HSL
	r, g, b, err := hexToRGB(baseHex)
	if err != nil {
		return nil, err
	}
	h, s, l := rgbToHsl(r, g, b)

	// 2) Lightness offsets around the base (index 4 == 0 offset).
	// Tuned to feel like the provided strip: slightly tighter steps
	// close to the base, larger steps toward the ends.
	// (Negative = darker than base, Positive = lighter than base)
	deltas := []float64{
		-0.35, -0.26, -0.18, -0.09, 0.00, +0.10, +0.20, +0.29, +0.38,
	}

	// 3) Build swatches with gentle saturation taper:
	// lighten ⇒ reduce S a bit; darken ⇒ nudge S up a touch.
	out := make([]string, len(deltas))
	for i, d := range deltas {
		// target lightness, clamped to [0,1]
		L := mustClamp01(l + d)

		// saturation adjustment: distance from the base index (4)
		dist := math.Abs(float64(i)-4.0) / 4.0 // 0 at base, up to 1 at ends
		sAdj := s*(1.0-0.15*dist) + 0.08*dist  // reduce S when lighter, add a little when darker
		sAdj = mustClamp01(sAdj)

		R, G, B := hslToRgb(h, sAdj, L)
		out[i] = rgbToHex(R, G, B)
	}

	return out, nil
}
