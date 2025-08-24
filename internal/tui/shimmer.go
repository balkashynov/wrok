package tui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// ShimmerConfig holds configuration for shimmer effects
type ShimmerConfig struct {
	Enabled           bool          // animations: on|off
	ReduceMotion      bool          // if true → use static highlight
	SpeedMs           int           // shimmer_speed_ms (default 100)
	WidthRatio        float64       // shimmer_width_ratio (default 0.25)
	CycleMs           int           // total cycle time in ms (default 1800)
	PauseBetweenMs    int           // pause between cycles in ms (default 500)
}

// ShimmerState holds the current state of a shimmer effect
type ShimmerState struct {
	Center       float64   // current center position
	LastUpdate   time.Time // last tick time
	Active       bool      // whether shimmer is active
	VisibleLen   int       // length of visible text
	Config       ShimmerConfig
	SupportsTrueColor bool
	IsPaused     bool      // whether we're in pause between cycles
	PauseStartTime time.Time // when the current pause started
}

// DefaultShimmerConfig returns default shimmer configuration
func DefaultShimmerConfig() ShimmerConfig {
	return ShimmerConfig{
		Enabled:        true,
		ReduceMotion:   false,
		SpeedMs:        100,
		WidthRatio:     0.25,
		CycleMs:        1800,
		PauseBetweenMs: 500,
	}
}

// NewShimmerState creates a new shimmer state
func NewShimmerState(config ShimmerConfig) *ShimmerState {
	return &ShimmerState{
		Center:       0,
		LastUpdate:   time.Now(),
		Active:       config.Enabled && !config.ReduceMotion,
		Config:       config,
		SupportsTrueColor: supportsShimmerTrueColor(),
	}
}

// supportsShimmerTrueColor detects if terminal supports truecolor for shimmer
func supportsShimmerTrueColor() bool {
	colorTerm := os.Getenv("COLORTERM")
	return colorTerm == "truecolor"
}

// Update advances the shimmer animation
func (s *ShimmerState) Update(visibleLen int) {
	if !s.Active || s.Config.ReduceMotion {
		return
	}
	
	now := time.Now()
	elapsed := now.Sub(s.LastUpdate)
	
	// Only update if enough time has passed
	if elapsed.Milliseconds() < int64(s.Config.SpeedMs) {
		return
	}
	
	s.VisibleLen = visibleLen
	if visibleLen <= 0 {
		return
	}
	
	// Handle pause between cycles
	if s.IsPaused {
		pauseElapsed := now.Sub(s.PauseStartTime)
		if pauseElapsed.Milliseconds() >= int64(s.Config.PauseBetweenMs) {
			// End pause, start new cycle
			s.IsPaused = false
			s.Center = -float64(visibleLen) * s.Config.WidthRatio // Start before the beginning
		}
		s.LastUpdate = now
		return
	}
	
	// Calculate velocity: how many glyphs to advance per tick
	ticksPerCycle := float64(s.Config.CycleMs) / float64(s.Config.SpeedMs)
	// Allow shimmer to travel beyond the text (start before, end after)
	totalDistance := float64(visibleLen) * (1.0 + 2.0*s.Config.WidthRatio) // Add buffer on both sides
	velocity := totalDistance / ticksPerCycle
	
	// Advance center position
	s.Center += velocity
	
	// Check if cycle is complete (shimmer has passed completely beyond the text)
	maxCenter := float64(visibleLen) + float64(visibleLen)*s.Config.WidthRatio
	if s.Center >= maxCenter {
		// Start pause between cycles
		s.IsPaused = true
		s.PauseStartTime = now
		s.Center = maxCenter // Hold at end position during pause
	}
	
	s.LastUpdate = now
}

// Reset resets the shimmer position (call when selection changes)
func (s *ShimmerState) Reset() {
	s.Center = 0
	s.LastUpdate = time.Now()
	s.IsPaused = false
	s.PauseStartTime = time.Time{}
}

// SetActive enables/disables shimmer
func (s *ShimmerState) SetActive(active bool) {
	s.Active = active && s.Config.Enabled && !s.Config.ReduceMotion
}

// RenderShimmerText renders text with shimmer effect
func (s *ShimmerState) RenderShimmerText(text string, maxWidth int) string {
	// Truncate text to fit within maxWidth
	visibleText := text
	if len(text) > maxWidth {
		visibleText = text[:maxWidth-3] + "..."
	}
	
	textLen := len(visibleText)
	if textLen == 0 {
		return ""
	}
	
	// Update shimmer position
	s.Update(textLen)
	
	// If shimmer is not active, return static text
	if !s.Active {
		return renderStaticShimmerText(visibleText)
	}
	
	// If terminal doesn't support truecolor, use simpler effect
	if !s.SupportsTrueColor {
		return renderFallbackShimmerText(visibleText, s.Center, s.Config.WidthRatio)
	}
	
	// Render full truecolor shimmer
	return s.renderTrueColorShimmer(visibleText)
}

// renderTrueColorShimmer renders the full truecolor shimmer effect
func (s *ShimmerState) renderTrueColorShimmer(text string) string {
	var b strings.Builder
	textLen := len(text)
	
	// Base color: #B1B8C7 (rgb 177,184,199)
	baseR, baseG, baseB := 177, 184, 199
	
	// Highlight colors (light violet ramp): #EAE6FF → #C4B5FD → #A78BFA
	highlightR, highlightG, highlightB := 234, 230, 255 // #EAE6FF (lightest)
	
	// Calculate sigma (bell curve width)
	sigma := s.Config.WidthRatio * float64(textLen) / 2.0
	if sigma < 1.0 {
		sigma = 1.0
	}
	
	for i, char := range text {
		// Calculate weight using Gaussian bell curve
		dx := float64(i) - s.Center
		weight := math.Exp(-(dx * dx) / (2 * sigma * sigma))
		
		// Clamp weight to [0,1]
		if weight > 1.0 {
			weight = 1.0
		}
		if weight < 0.0 {
			weight = 0.0
		}
		
		// Linear blend: out = base*(1-w) + highlight*w
		finalR := int(float64(baseR)*(1-weight) + float64(highlightR)*weight)
		finalG := int(float64(baseG)*(1-weight) + float64(highlightG)*weight)
		finalB := int(float64(baseB)*(1-weight) + float64(highlightB)*weight)
		
		// Write the character with its color
		b.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", finalR, finalG, finalB, char))
	}
	
	// Reset color
	b.WriteString("\033[0m")
	
	return b.String()
}

// renderStaticShimmerText renders static highlighted text (no animation)
func renderStaticShimmerText(text string) string {
	// Use a static accent color for reduced motion
	return fmt.Sprintf("\033[38;2;167;139;250m%s\033[0m", text) // ColorAccentBright
}

// renderFallbackShimmerText renders a simple fallback shimmer for non-truecolor terminals
func renderFallbackShimmerText(text string, center float64, widthRatio float64) string {
	textLen := len(text)
	if textLen == 0 {
		return text
	}
	
	// Simple fallback: highlight a few characters around the center
	highlightWidth := int(widthRatio * float64(textLen))
	if highlightWidth < 1 {
		highlightWidth = 1
	}
	
	startHighlight := int(center) - highlightWidth/2
	endHighlight := startHighlight + highlightWidth
	
	var b strings.Builder
	for i, char := range text {
		if i >= startHighlight && i < endHighlight {
			// Highlight color (256-color approximation of our purple)
			b.WriteString(fmt.Sprintf("\033[38;5;147m%c", char)) // Light purple
		} else {
			// Base color (256-color approximation)
			b.WriteString(fmt.Sprintf("\033[38;5;250m%c", char)) // Light grey
		}
	}
	
	// Reset color
	b.WriteString("\033[0m")
	
	return b.String()
}

// GetTickInterval returns the interval for tea.Tick commands
func (s *ShimmerState) GetTickInterval() time.Duration {
	if !s.Active {
		return 0 // No ticking needed
	}
	return time.Duration(s.Config.SpeedMs) * time.Millisecond
}

// ShouldTick returns true if shimmer should be ticking
func (s *ShimmerState) ShouldTick() bool {
	return s.Active && s.Config.Enabled && !s.Config.ReduceMotion
}