package gallery

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

// leftRedRightBlue makes a w×h image: left half red, right half blue.
func leftRedRightBlue(w, h int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.RGBA{255, 0, 0, 255}
			if x >= w/2 {
				c = color.RGBA{0, 0, 255, 255}
			}
			im.SetRGBA(x, y, c)
		}
	}
	return im
}

func encodeJPEG(t *testing.T, im image.Image) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := jpeg.Encode(&b, im, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// withExif inserts an APP1 Exif segment (little-endian TIFF) holding Orientation and a fake
// GPS marker string right after the JPEG SOI.
func withExif(jpg []byte, orientation uint16) []byte {
	var tiff bytes.Buffer
	tiff.WriteString("II")
	_ = binary.Write(&tiff, binary.LittleEndian, uint16(42))
	_ = binary.Write(&tiff, binary.LittleEndian, uint32(8))
	_ = binary.Write(&tiff, binary.LittleEndian, uint16(1)) // one IFD entry
	_ = binary.Write(&tiff, binary.LittleEndian, uint16(0x0112))
	_ = binary.Write(&tiff, binary.LittleEndian, uint16(3)) // SHORT
	_ = binary.Write(&tiff, binary.LittleEndian, uint32(1))
	_ = binary.Write(&tiff, binary.LittleEndian, orientation)
	_ = binary.Write(&tiff, binary.LittleEndian, uint16(0))
	_ = binary.Write(&tiff, binary.LittleEndian, uint32(0)) // no next IFD
	tiff.WriteString("GPSLatitude=43.6150,-116.2023")
	payload := append([]byte("Exif\x00\x00"), tiff.Bytes()...)
	seg := []byte{0xFF, 0xE1, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}
	out := append([]byte{}, jpg[:2]...)
	out = append(out, seg...)
	out = append(out, payload...)
	return append(out, jpg[2:]...)
}

func decode(t *testing.T, b []byte) image.Image {
	t.Helper()
	im, format, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" {
		t.Fatalf("output format %q, want jpeg", format)
	}
	return im
}

func isReddish(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r > 0xB000 && g < 0x5000 && b < 0x5000
}

func isBluish(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return b > 0xB000 && r < 0x5000 && g < 0x5000
}

func TestProcessDownscalesAndMakesThumb(t *testing.T) {
	res, err := Process(encodeJPEG(t, leftRedRightBlue(3200, 2000)))
	if err != nil {
		t.Fatal(err)
	}
	if res.DisplayWidth != 1600 || res.DisplayHeight != 1000 {
		t.Errorf("display %dx%d, want 1600x1000", res.DisplayWidth, res.DisplayHeight)
	}
	if res.ThumbWidth != ThumbMaxEdge || res.ThumbHeight != 400 {
		t.Errorf("thumb %dx%d, want %dx400", res.ThumbWidth, res.ThumbHeight, ThumbMaxEdge)
	}
	d := decode(t, res.Display)
	if d.Bounds().Dx() != res.DisplayWidth || d.Bounds().Dy() != res.DisplayHeight {
		t.Error("reported size doesn't match encoded display image")
	}
	if len(res.Hash) != 12 {
		t.Errorf("hash %q", res.Hash)
	}
}

func TestProcessNeverUpscales(t *testing.T) {
	res, err := Process(encodeJPEG(t, leftRedRightBlue(300, 200)))
	if err != nil {
		t.Fatal(err)
	}
	if res.DisplayWidth != 300 || res.DisplayHeight != 200 || res.ThumbWidth != 300 {
		t.Errorf("got display %dx%d thumb %d", res.DisplayWidth, res.DisplayHeight, res.ThumbWidth)
	}
}

func TestProcessAppliesExifRotationAndStripsMetadata(t *testing.T) {
	// Orientation 6 = the camera was rotated; viewers turn it 90° clockwise. The left (red)
	// half of the stored pixels must end up on top.
	src := withExif(encodeJPEG(t, leftRedRightBlue(200, 100)), 6)
	if got := jpegOrientation(src); got != 6 {
		t.Fatalf("parsed orientation %d, want 6", got)
	}
	res, err := Process(src)
	if err != nil {
		t.Fatal(err)
	}
	if res.DisplayWidth != 100 || res.DisplayHeight != 200 {
		t.Fatalf("rotated size %dx%d, want 100x200", res.DisplayWidth, res.DisplayHeight)
	}
	im := decode(t, res.Display)
	if !isReddish(im.At(50, 20)) || !isBluish(im.At(50, 180)) {
		t.Errorf("unexpected colors after rotation: top=%v bottom=%v", im.At(50, 20), im.At(50, 180))
	}
	for _, out := range [][]byte{res.Display, res.Thumb} {
		if bytes.Contains(out, []byte("Exif")) || bytes.Contains(out, []byte("GPS")) {
			t.Error("output still contains EXIF/GPS metadata")
		}
	}
}

func TestOrientationVariants(t *testing.T) {
	for o := 1; o <= 8; o++ {
		res, err := Process(withExif(encodeJPEG(t, leftRedRightBlue(80, 40)), uint16(o)))
		if err != nil {
			t.Fatalf("orientation %d: %v", o, err)
		}
		wantW, wantH := 80, 40
		if o >= 5 {
			wantW, wantH = 40, 80
		}
		if res.DisplayWidth != wantW || res.DisplayHeight != wantH {
			t.Errorf("orientation %d: %dx%d, want %dx%d", o, res.DisplayWidth, res.DisplayHeight, wantW, wantH)
		}
	}
}

func TestProcessFlattensTransparentPNGOntoWhite(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 50, 50)) // fully transparent
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	res, err := Process(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	r, g, bl, _ := decode(t, res.Display).At(25, 25).RGBA()
	if r < 0xF000 || g < 0xF000 || bl < 0xF000 {
		t.Errorf("transparent area should become white, got %x %x %x", r, g, bl)
	}
}

func TestProcessWebP(t *testing.T) {
	data, err := os.ReadFile("testdata/wide.webp")
	if err != nil {
		t.Fatal(err)
	}
	res, err := Process(data)
	if err != nil {
		t.Fatal(err)
	}
	im := decode(t, res.Display)
	if res.DisplayWidth != 40 || !isReddish(im.At(5, 10)) || !isBluish(im.At(35, 10)) {
		t.Errorf("webp decode wrong: %dx%d", res.DisplayWidth, res.DisplayHeight)
	}
}

func TestProcessRejects(t *testing.T) {
	heic := append([]byte{0, 0, 0, 0x18}, []byte("ftypheic\x00\x00\x00\x00mif1heic")...)
	gif := []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;")
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{"heic", heic, ErrHEIC},
		{"gif", gif, ErrUnsupported},
		{"text", []byte("hello, not an image"), ErrUnsupported},
		{"truncated jpeg", encodeJPEG(t, leftRedRightBlue(64, 64))[:40], ErrCorrupt},
		{"too large", append([]byte{0xFF, 0xD8, 0xFF}, make([]byte, MaxUploadBytes)...), ErrTooLarge},
	}
	for _, c := range cases {
		if _, err := Process(c.data); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

func TestProcessRejectsHugeDimensions(t *testing.T) {
	// A tiny PNG file that claims 10000×10000 pixels: rejected from the header alone.
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewGray(image.Rect(0, 0, 10000, 10000))); err != nil {
		t.Fatal(err)
	}
	if _, err := Process(b.Bytes()); !errors.Is(err, ErrTooManyPx) {
		t.Errorf("err = %v, want ErrTooManyPx", err)
	}
}
