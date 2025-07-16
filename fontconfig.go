package mailprint

// #cgo pkg-config: fontconfig
// #include <fontconfig/fontconfig.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// FindFont uses Fontconfig to find the path to a font file given a
// string-formatted Fontconfig pattern (a font name) and weight.
//
// It returns the path of the font file, discarding additional
// properties that may have been provided.
//
// The Fontconfig pattern syntax ("Font name") is described at:
// https://www.freedesktop.org/software/fontconfig/fontconfig-user.html#AEN36
func FindFont(fontName string, bold bool) (string, error) {
	// Initialize fontconfig
	if C.FcInit() == C.FcFalse {
		return "", fmt.Errorf("failed to initialize fontconfig")
	}

	// Parse the font name
	cFontName := C.CString(fontName)
	defer C.free(unsafe.Pointer(cFontName))
	pattern := C.FcNameParse((*C.uchar)(unsafe.Pointer(cFontName)))
	if pattern == nil {
		return "", fmt.Errorf("failed to parse font name: %s", fontName)
	}
	defer C.FcPatternDestroy(pattern)

	if bold {
		cWeight := C.CString("weight")
		defer C.free(unsafe.Pointer(cWeight))
		C.FcPatternAddInteger(pattern, cWeight, C.FC_WEIGHT_BOLD)
	}

	// Configure the pattern
	C.FcConfigSubstitute(nil, pattern, C.FcMatchPattern)
	C.FcDefaultSubstitute(pattern)

	// Find the font
	var result C.FcResult
	match := C.FcFontMatch(nil, pattern, &result)
	if match == nil {
		return "", fmt.Errorf("no font found for: %s", fontName)
	}
	defer C.FcPatternDestroy(match)

	// Get the font file path
	var cFilePath *C.char
	cFileString := C.CString("file")
	defer C.free(unsafe.Pointer(cFileString))
	if C.FcPatternGetString(match, cFileString, 0, (**C.uchar)(unsafe.Pointer(&cFilePath))) != C.FcResultMatch {
		return "", fmt.Errorf("failed to get font file path for: %s", fontName)
	}

	return C.GoString(cFilePath), nil
}

func FindFonts(headerFontName, contentFontName string) (FontPathOptions, error) {
	h, err := FindFont(headerFontName, false)
	if err != nil {
		return FontPathOptions{}, fmt.Errorf("looking up header font %q: %w", headerFontName, err)
	}
	hb, err := FindFont(headerFontName, true)
	if err != nil {
		return FontPathOptions{}, fmt.Errorf("looking up bold header font %q: %w", headerFontName, err)
	}
	c, err := FindFont(contentFontName, false)
	if err != nil {
		return FontPathOptions{}, fmt.Errorf("looking up content font %q: %w", contentFontName, err)
	}
	return FontPathOptions{
		Header:     h,
		HeaderBold: hb,
		Content:    c,
	}, nil
}
