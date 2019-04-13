package bmfonter

import (
	"image"
	"image/draw"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/KeKsBoTer/gofnt"
)

type Font struct {
	Font     gofnt.Font
	Chars    map[int]gofnt.Char
	Image    image.Image
	Subfonts []Font
}

func InitFont(fontPath, imgPath string) (font Font, err error) {
	// Read fnt data.
	content, err := ioutil.ReadFile(filepath.FromSlash(fontPath))
	if err != nil {
		return
	}
	f, err := gofnt.Parse(string(content))
	if err != nil {
		return
	}

	// Read image data.
	imgFile, err := os.Open(filepath.FromSlash(imgPath))
	if err != nil {
		return
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return
	}

	font.Font = *f
	font.Chars = make(map[int]gofnt.Char)
	for _, c := range f.Chars {
		font.Chars[int(c.Id)] = c
	}
	font.Image = img

	return font, nil
}

func (font *Font) AddSubFont(fontPath, imgPath string) error {
	// Read fnt data.
	content, err := ioutil.ReadFile(filepath.FromSlash(fontPath))
	if err != nil {
		return err
	}
	f, err := gofnt.Parse(string(content))
	if err != nil {
		return err
	}

	// Read image data.
	imgFile, err := os.Open(filepath.FromSlash(imgPath))
	if err != nil {
		return err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	subfont := Font{}
	subfont.Font = *f
	subfont.Chars = make(map[int]gofnt.Char)
	for _, c := range f.Chars {
		subfont.Chars[int(c.Id)] = c
	}
	subfont.Image = img

	font.Subfonts = append(font.Subfonts, subfont)

	return nil
}

func (f *Font) RenderChar(dst draw.Image, x, y int, r rune) int {
	font := *f
	for _, sf := range f.Subfonts {
		if _, ok := sf.Chars[int(r)]; ok {
			font = sf
		}
	}
	c := font.Chars[int(r)]
	draw.Draw(dst, image.Rectangle{image.Point{x + c.XOffset, y + c.YOffset}, image.Point{x + c.Width + c.XOffset, y + c.Height + c.YOffset}}, font.Image, image.Point{c.X, c.Y}, draw.Over)
	return c.XAdvanced
}

func (f *Font) RenderString(dst draw.Image, x, y int, s string) int {
	for _, r := range s {
		x += f.RenderChar(dst, x, y, r)
	}
	return x
}

func (f *Font) RenderTextBox(dst draw.Image, x, y int, width, height int, centeredX, centeredY bool, s string) {
	// Split string into lines that fit into the bounding box.
	var lines []string
	var line, word string
	lineWidth := 0
	words := strings.Fields(s)
	font := *f
	for len(words) > 0 {
		word, words = words[0], words[1:]
		wordWidth := 0
		for _, char := range word {
			for _, sf := range f.Subfonts {
				if _, ok := sf.Chars[int(char)]; ok {
					font = sf
				}
			}
			wordWidth += font.Chars[int(char)].XAdvanced
			font = *f
		}
		// Add spacebar width if this is not the first or last word.
		if lineWidth != 0 || len(words) > 0 {
			wordWidth += font.Chars[int(32)].XAdvanced
		}
		if lineWidth+wordWidth < width {
			lineWidth += wordWidth
			if len(line) > 0 {
				line += " "
			}
			line += word
			if len(words) == 0 {
				lines = append(lines, line)
			}
		} else {
			words = append([]string{word}, words...)
			lines = append(lines, line)
			line = ""
			lineWidth = 0
			if len(lines)*font.Font.Common.LineHeight > height {
				break
			}
		}
	}

	if centeredY {
		y -= len(lines) * font.Font.Common.LineHeight / 2
	}

	for _, line := range lines {
		x1 := x
		if centeredX {
			xWidth := 0
			font = *f
			for _, r := range line {
				for _, sf := range f.Subfonts {
					if _, ok := sf.Chars[int(r)]; ok {
						font = sf
					}
				}
				xWidth += font.Chars[int(r)].XAdvanced
			}
			x1 -= xWidth / 2
		}
		f.RenderString(dst, x1, y, line)
		y += f.Font.Common.LineHeight
	}
}
