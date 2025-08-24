package tui

// Color constants for wrok TUI theme
const (
	// Base Colors
	ColorAppBackground  = "" // Use terminal default background
	ColorCardBackground = "#1B1530" // Dark purple
	ColorBorder         = "#3A3F55" // Grey-blue

	// Text Colors
	ColorPrimaryText   = "#E6EAF2" // Primary text (field labels, user input, titles)
	ColorSecondaryText = "#B1B8C7" // Secondary text - subtle purple-tinted grey
	ColorDisabledText  = "#6D7383" // Disabled/muted text
	ColorPlaceholder   = "#B1B8C7" // Same as secondary - that beautiful purple-grey with shine
	ColorHelpText      = "240"      // Dark grey for help text

	// Accent Colors (Purple theme)
	ColorAccentMain   = "#7C3AED" // Logo, accent elements, active borders
	ColorAccentBright = "#A78BFA" // Hover, highlights, current step

	// State Colors
	ColorError   = "#EF4444" // Validation errors
	ColorSuccess = "#22C55E" // Success, confirmations
	ColorWarning = "#F59E0B" // Warnings
)