package ascii

import (
	"fmt"
	"os"
)

// CalibrateTerminal performs interactive calibration of terminal display
func CalibrateTerminal() error {
	fmt.Println("Terminal Calibration")
	fmt.Println("===================")
	fmt.Println()
	fmt.Println("This tool helps calibrate your terminal's display of Unicode characters.")
	fmt.Println("Follow the prompts to ensure proper rendering.")
	fmt.Println()

	// Test box drawing
	fmt.Println("Box Drawing Test:")
	box := DrawBox("Hello, World!", 20)
	fmt.Print(box)
	fmt.Println()

	// Test string width
	testStrings := []string{
		"Hello",
		"Hello World",
		"Café",
		"🚀 Rocket",
	}

	fmt.Println("String Width Test:")
	for _, s := range testStrings {
		width := StringWidth(s)
		fmt.Printf("'%s' -> width: %d\n", s, width)
	}
	fmt.Println()

	fmt.Println("Calibration complete. If the output looks correct, your terminal is properly configured.")
	fmt.Println("If characters appear garbled, you may need to install a Unicode font or adjust your terminal settings.")

	return nil
}

// RunCalibration is a convenience function to run calibration and exit
func RunCalibration() {
	if err := CalibrateTerminal(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during calibration: %v\n", err)
		os.Exit(1)
	}
}
