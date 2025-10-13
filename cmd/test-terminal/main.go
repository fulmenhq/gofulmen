package main

import (
	"fmt"
	"os"

	"github.com/fulmenhq/gofulmen/ascii"
)

func main() {
	fmt.Println("Testing terminal override rendering...")
	fmt.Println("Current TERM_PROGRAM:", os.Getenv("TERM_PROGRAM"))
	fmt.Println()

	config := ascii.GetTerminalConfig()
	if config != nil {
		fmt.Printf("Detected terminal: %s\n", config.Name)
		fmt.Printf("Override count: %d\n", len(config.Overrides))
		fmt.Println()
	} else {
		fmt.Println("No terminal config detected (using defaults)")
		fmt.Println()
	}

	testEmojis := []string{
		"⏱️", "☠️", "☹️", "⚠️", "✌️",
		"🎗️", "🎟️", "🖐️", "🛠️", "ℹ️",
	}

	fmt.Println("Testing emoji rendering:")
	for _, emoji := range testEmojis {
		width := ascii.StringWidth(emoji)
		line := fmt.Sprintf("%s Width: %d", emoji, width)
		lineWidth := ascii.StringWidth(line)
		fmt.Printf("%s (line width: %d)\n", line, lineWidth)
	}

	fmt.Println()
	testBox := ascii.DrawBox("⚠️  Important Warning ☠️", 30)
	fmt.Println("Test box with emojis:")
	fmt.Println(testBox)
}
