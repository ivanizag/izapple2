package screen

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
)

// TestScenario is the computer video state
type TestScenario struct {
	VideoMode     uint32     `json:"mode"`
	VideoModeName string     `json:"name"`
	ScreenModes   []int      `json:"screens"`
	TextPages     [4][]uint8 `json:"text"`
	VideoPages    [4][]uint8 `json:"video"`
	SVideoPage    []uint8    `json:"svideo"`
}

func grabTestScenario(vs VideoSource) *TestScenario {
	textPages := [4][]uint8{
		cloneSlice(vs.GetTextMemory(false, false)),
		cloneSlice(vs.GetTextMemory(false, true)),
		cloneSlice(vs.GetTextMemory(true, false)),
		cloneSlice(vs.GetTextMemory(true, true)),
	}
	videoPages := [4][]uint8{
		cloneSlice(vs.GetVideoMemory(false, false)),
		cloneSlice(vs.GetVideoMemory(false, true)),
		cloneSlice(vs.GetVideoMemory(true, false)),
		cloneSlice(vs.GetVideoMemory(true, true)),
	}

	knownModes := []int{
		ScreenModeGreen,
		ScreenModeNTSC,
		ScreenModePlain,
	}

	return &TestScenario{
		vs.GetCurrentVideoMode(),
		VideoModeName(vs),
		knownModes,
		textPages,
		videoPages,
		cloneSlice(vs.GetSuperVideoMemory()),
	}
}

func cloneSlice(src []uint8) []uint8 {
	dst := make([]uint8, len(src))
	copy(dst, src)
	return dst
}

func loadTestScenario(filename string) (*TestScenario, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ts TestScenario
	err = json.Unmarshal(bytes, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func (ts *TestScenario) save(dir string) (string, error) {
	bytes, err := json.Marshal(ts)
	if err != nil {
		return "", err
	}

	pattern := fmt.Sprintf("%v_*.json", strings.ToLower(ts.VideoModeName))
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(bytes)
	return file.Name(), err
}

// GetCurrentVideoMode returns the active video mode
func (ts *TestScenario) GetCurrentVideoMode() uint32 {
	return ts.VideoMode
}

func optionsToIndex(secondPage bool, ext bool) int {
	index := 0
	if secondPage {
		index += 2
	}
	if ext {
		index++
	}
	return index
}

// GetTextMemory returns a slice to the text memory pages
func (ts *TestScenario) GetTextMemory(secondPage bool, ext bool) []uint8 {
	return ts.TextPages[optionsToIndex(secondPage, ext)]
}

// GetVideoMemory returns a slice to the video memory pages
func (ts *TestScenario) GetVideoMemory(secondPage bool, ext bool) []uint8 {
	return ts.VideoPages[optionsToIndex(secondPage, ext)]
}

// GetCharacterPixel returns the pixel as output by the character generator
func (ts *TestScenario) GetCharacterPixel(char uint8, rowInChar int, colInChar int, isAltText bool, isFlashedFrame bool) bool {
	// We don't have a character generator. We will return a square or blank for spaces
	if char&0x3f == 0x20 {
		return false // Space char
	}

	return !(rowInChar == 0 || rowInChar == 7 || colInChar == 0 || colInChar == 6)
}

// GetSuperVideoMemory returns a slice to the SHR video memory
func (ts *TestScenario) GetSuperVideoMemory() []uint8 {
	return ts.SVideoPage
}

// GetCardImage returns an image provided by a card, like the videx card
func (ts *TestScenario) GetCardImage(light color.Color) *image.RGBA {
	return nil
}

// SupportsLowercase returns true if the video source supports lowercase
func (ts *TestScenario) SupportsLowercase() bool {
	return true
}

func buildImageName(name string, screenMode int, altSet bool) string {
	var screenName string
	switch screenMode {
	case ScreenModeGreen:
		screenName = "green"
	case ScreenModeNTSC:
		screenName = "ntsc"
	case ScreenModePlain:
		screenName = "plain"
	default:
		screenName = "unknown"
	}

	if altSet {
		screenName += "_new"
	}

	return strings.TrimSuffix(name, ".json") +
		screenName + ".png"
}

func (ts *TestScenario) generateSnapshots(baseName string, altSet bool) error {
	for _, screen := range ts.ScreenModes {
		image := Snapshot(ts, screen)
		imageName := buildImageName(baseName, screen, altSet)
		f, err := os.Create(imageName)
		if err != nil {
			return err
		}
		defer f.Close()

		png.Encode(f, image)
	}
	return nil
}

// AddScenario Generate a new video scenario
func AddScenario(vs VideoSource, dir string) error {
	// Get memory contents
	ts := grabTestScenario(vs)

	name, err := ts.save(dir)
	if err != nil {
		return err
	}

	return ts.generateSnapshots(name, false)
}
