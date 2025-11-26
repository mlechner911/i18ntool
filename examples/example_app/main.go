package main

import (
	"fmt"
	// pretend i18n loader
)

// main is a small demo that prints referenced translation keys.
func main() {
	// This is a demo app that references some translation keys from the examples locales.
	// It intentionally does not compile — it's just a usage demo.

	// Example usage (pseudo-code):
	// msg := i18n.T("common.greeting.hello")
	// fmt.Println(msg)

	// References (some keys used, some missing to demonstrate detection):
	fmt.Println("Using keys:")
	fmt.Println("common.greeting.hello")
	fmt.Println("common.button.save")
	fmt.Println("errors.network.timeout")

	// Unused / missing example keys (exist in locales but not used here):
	// common.greeting.welcome
	// user.profile.name

	// Invalid use to make it non-compilable (uncomment to see compile error):
	// var x = nonExistingFunction() // intentional error
}
