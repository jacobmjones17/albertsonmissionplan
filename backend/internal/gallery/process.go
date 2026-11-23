// Package gallery turns uploaded drawing scans into web-sized JPEGs: it validates the file
// type, applies the camera's EXIF rotation, flattens transparency onto white, downscales to a
// display size and a thumbnail, and re-encodes. Re-encoding writes only pixels, so EXIF/GPS
// and every other metadata block from the original file are dropped.
package gallery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // register PNG decoder
	"net/http"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register WebP decoder
)

const (
	// MaxUploadBytes is the largest scan accepted (phone photos and 300–600 dpi scans fit).
	MaxUploadBytes = 15 << 20
	// MaxPixels guards against decompression bombs (a 600 dpi letter-size scan is ~34 MP).
	MaxPixels = 60_000_000
	// DisplayMaxEdge is the long edge of the image shown in the lightbox.
	DisplayMaxEdge = 1600
	// ThumbMaxEdge is the long edge of the grid thumbnail (sharp at 2x on ~300px columns).
	ThumbMaxEdge = 640

	displayQuality = 85
	thumbQuality   = 80
)

var (
	ErrTooLarge    = fmt.Errorf("That file is larger than %d MB. Try scanning at 300 dpi or taking a photo instead.", MaxUploadBytes>>20)
	ErrUnsupported = errors.New("Please upload a JPEG, PNG, or WebP image.")
	ErrHEIC        = errors.New("HEIC photos can't be read yet. On iPhone, set Settings → Camera → Formats → Most Compatible, or upload from the Photos picker (it converts to JPEG).")
	ErrTooManyPx   = errors.New("That image has too many pixels. Please scan at 300 dpi or lower.")
	ErrCorrupt     = errors.New("That image couldn't be read. It may be damaged; try exporting it again.")
)

// Result holds the processed images.
type Result struct {
	Display       []byte
	DisplayWidth  int
	DisplayHeight int
	Thumb         []byte
	ThumbWidth    int
	ThumbHeight   int
	// Hash is a short content hash of Display, used to version cache-forever image URLs.
	Hash string
}

// Sniff returns the detected image type ("jpeg", "png", "webp") or an error for anything else.
func Sniff(data []byte) (string, error) {
	if isHEIC(data) {
		return "", ErrHEIC
	}
	switch http.DetectContentType(data) {
	case "image/jpeg":
		return "jpeg", nil
	case "image/png":
		return "png", nil
	case "image/webp":
		return "webp", nil
	}
	return "", ErrUnsupported
}

func isHEIC(b []byte) bool {
	if len(b) < 12 || string(b[4:8]) != "ftyp" {
		return false
	}
	switch string(b[8:12]) {
	case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1", "avif":
		return true
	}
	return false
}

// decodeUpload validates an upload (size, type, pixel count) and decodes it, returning the image and
// its EXIF orientation (1 when not a JPEG or absent).
func decodeUpload(data []byte) (image.Image, int, error) {
	if len(data) > MaxUploadBytes {
		return nil, 0, ErrTooLarge
	}
	kind, err := Sniff(data)
	if err != nil {
		return nil, 0, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, ErrCorrupt
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, 0, ErrCorrupt
	}
	if cfg.Width*cfg.Height > MaxPixels {
		return nil, 0, ErrTooManyPx
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, ErrCorrupt
	}
	orientation := 1
	if kind == "jpeg" {
		orientation = jpegOrientation(data)
	}
	return src, orientation, nil
}

// ProcessCover validates one upload and center-crops it to fill exactly w×h (display) and
// thumbW×thumbH (thumbnail), like CSS object-fit: cover. Used for artwork that must keep a
// fixed shape, such as the 3:2 Monthly Challenges cards. Metadata is dropped the same way as
// Process (only pixels are re-encoded).
func ProcessCover(data []byte, w, h, thumbW, thumbH int) (*Result, error) {
	src, orientation, err := decodeUpload(data)
	if err != nil {
		return nil, err
	}
	// Downscale first (bounded memory), then rotate upright, then crop+scale to the target.
	work := orient(scaleOnWhite(src, 2*maxInt(w, h), draw.CatmullRom), orientation)
	display := coverScale(work, w, h, draw.CatmullRom)
	thumb := coverScale(display, thumbW, thumbH, draw.ApproxBiLinear)
	var db, tb bytes.Buffer
	if err := jpeg.Encode(&db, display, &jpeg.Options{Quality: displayQuality}); err != nil {
		return nil, err
	}
	if err := jpeg.Encode(&tb, thumb, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(db.Bytes())
	return &Result{
		Display: db.Bytes(), DisplayWidth: w, DisplayHeight: h,
		Thumb: tb.Bytes(), ThumbWidth: thumbW, ThumbHeight: thumbH,
		Hash: hex.EncodeToString(sum[:6]),
	}, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// coverScale crops the centre of src to the w:h aspect ratio and scales it to exactly w×h.
func coverScale(src image.Image, w, h int, scaler draw.Scaler) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	crop := b
	if sw*h > sh*w { // too wide: trim the sides
		cw := sh * w / h
		x0 := b.Min.X + (sw-cw)/2
		crop = image.Rect(x0, b.Min.Y, x0+cw, b.Max.Y)
	} else if sw*h < sh*w { // too tall: trim top and bottom
		ch := sw * h / w
		y0 := b.Min.Y + (sh-ch)/2
		crop = image.Rect(b.Min.X, y0, b.Max.X, y0+ch)
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	scaler.Scale(dst, dst.Bounds(), src, crop, draw.Over, nil)
	return dst
}

// Process validates and converts one upload.
func Process(data []byte) (*Result, error) {
	src, orientation, err := decodeUpload(data)
	if err != nil {
		return nil, err
	}

	// Scale straight from the decoded source into a small white canvas (flattens PNG/WebP
	// transparency and avoids a full-resolution RGBA copy, which matters on 512 MB hosts).
	display := orient(scaleOnWhite(src, DisplayMaxEdge, draw.CatmullRom), orientation)
	thumb := scaleOnWhite(display, ThumbMaxEdge, draw.ApproxBiLinear)

	var db, tb bytes.Buffer
	if err := jpeg.Encode(&db, display, &jpeg.Options{Quality: displayQuality}); err != nil {
		return nil, err
	}
	if err := jpeg.Encode(&tb, thumb, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(db.Bytes())
	return &Result{
		Display:       db.Bytes(),
		DisplayWidth:  display.Bounds().Dx(),
		DisplayHeight: display.Bounds().Dy(),
		Thumb:         tb.Bytes(),
		ThumbWidth:    thumb.Bounds().Dx(),
		ThumbHeight:   thumb.Bounds().Dy(),
		Hash:          hex.EncodeToString(sum[:6]),
	}, nil
}

// fit returns w×h scaled so the long edge is at most maxEdge (never upscales).
func fit(w, h, maxEdge int) (int, int) {
	long := w
	if h > long {
		long = h
	}
	if long <= maxEdge {
		return w, h
	}
	nw := int(float64(w)*float64(maxEdge)/float64(long) + 0.5)
	nh := int(float64(h)*float64(maxEdge)/float64(long) + 0.5)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return nw, nh
}

func scaleOnWhite(src image.Image, maxEdge int, scaler draw.Scaler) *image.RGBA {
	b := src.Bounds()
	w, h := fit(b.Dx(), b.Dy(), maxEdge)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	scaler.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

// orient applies an EXIF orientation (1–8) so the picture appears upright.
func orient(src *image.RGBA, o int) *image.RGBA {
	if o < 2 || o > 8 {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if o >= 5 {
		dw, dh = h, w
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var nx, ny int
			switch o {
			case 2: // mirror horizontal
				nx, ny = w-1-x, y
			case 3: // rotate 180
				nx, ny = w-1-x, h-1-y
			case 4: // mirror vertical
				nx, ny = x, h-1-y
			case 5: // transpose
				nx, ny = y, x
			case 6: // rotate 90 CW
				nx, ny = h-1-y, x
			case 7: // transverse
				nx, ny = h-1-y, w-1-x
			case 8: // rotate 90 CCW
				nx, ny = y, w-1-x
			}
			dst.SetRGBA(nx, ny, src.RGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

// jpegOrientation reads the EXIF Orientation tag (0x0112) from a JPEG's APP1 segment.
// Returns 1 (normal) when absent or unreadable.
func jpegOrientation(b []byte) int {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return 1
	}
	i := 2
	for i+4 <= len(b) {
		if b[i] != 0xFF {
			return 1
		}
		marker := b[i+1]
		if marker == 0xDA || marker == 0xD9 { // start of scan / end: no more metadata
			return 1
		}
		segLen := int(b[i+2])<<8 | int(b[i+3])
		if segLen < 2 || i+2+segLen > len(b) {
			return 1
		}
		seg := b[i+4 : i+2+segLen]
		if marker == 0xE1 && len(seg) > 14 && string(seg[:6]) == "Exif\x00\x00" {
			return tiffOrientation(seg[6:])
		}
		i += 2 + segLen
	}
	return 1
}

func tiffOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var u16 func([]byte) int
	var u32 func([]byte) int
	switch string(t[:2]) {
	case "II":
		u16 = func(p []byte) int { return int(p[0]) | int(p[1])<<8 }
		u32 = func(p []byte) int { return int(p[0]) | int(p[1])<<8 | int(p[2])<<16 | int(p[3])<<24 }
	case "MM":
		u16 = func(p []byte) int { return int(p[0])<<8 | int(p[1]) }
		u32 = func(p []byte) int { return int(p[0])<<24 | int(p[1])<<16 | int(p[2])<<8 | int(p[3]) }
	default:
		return 1
	}
	off := u32(t[4:8])
	if off < 8 || off+2 > len(t) {
		return 1
	}
	n := u16(t[off : off+2])
	for k := 0; k < n; k++ {
		e := off + 2 + k*12
		if e+12 > len(t) {
			return 1
		}
		if u16(t[e:e+2]) == 0x0112 {
			v := u16(t[e+8 : e+10])
			if v >= 1 && v <= 8 {
				return v
			}
			return 1
		}
	}
	return 1
}
