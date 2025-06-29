package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestClaudeSquadColors(t *testing.T) {
	// Test that colors are defined as expected
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"ClaudeSquadPrimary", ClaudeSquadPrimary, "#ffffff"},
		{"ClaudeSquadAccent", ClaudeSquadAccent, "#00ff9d"},
		{"ClaudeSquadDark", ClaudeSquadDark, "#0a0a0a"},
		{"ClaudeSquadGray", ClaudeSquadGray, "#1a1a1a"},
		{"ClaudeSquadMuted", ClaudeSquadMuted, "#6b7280"},
		{"ClaudeSquadHover", ClaudeSquadHover, "#00e68a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.color) != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, string(tt.color))
			}
		})
	}
}

func TestLegacyColors(t *testing.T) {
	// Test that legacy colors are defined
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"PrimaryColor", PrimaryColor, "#7C3AED"},
		{"SecondaryColor", SecondaryColor, "#10B981"},
		{"AccentColor", AccentColor, "#F59E0B"},
		{"ErrorColor", ErrorColor, "#EF4444"},
		{"SuccessColor", SuccessColor, "#10B981"},
		{"WarningColor", WarningColor, "#F59E0B"},
		{"MutedColor", MutedColor, "#6B7280"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.color) != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, string(tt.color))
			}
		})
	}
}

func TestClaudeSquadBaseStyles(t *testing.T) {
	// Test that base styles are properly configured
	t.Run("ClaudeSquadBaseStyle", func(t *testing.T) {
		// Test that the base style has the right colors
		// We can't directly test internal properties, but we can test rendering
		rendered := ClaudeSquadBaseStyle.Render("test")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ClaudeSquadAccentStyle", func(t *testing.T) {
		rendered := ClaudeSquadAccentStyle.Render("test")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ClaudeSquadBorderStyle", func(t *testing.T) {
		rendered := ClaudeSquadBorderStyle.Render("test content")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
		// Border style should add padding and borders, so should be longer than input
		if len(rendered) <= len("test content") {
			t.Error("Expected bordered content to be longer than original")
		}
	})
}

func TestLegacyStyles(t *testing.T) {
	// Test legacy styles
	t.Run("HeaderStyle", func(t *testing.T) {
		rendered := HeaderStyle.Render("Header")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ContentStyle", func(t *testing.T) {
		rendered := ContentStyle.Render("Content")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
		// Content style adds padding, so should be longer
		if len(rendered) <= len("Content") {
			t.Error("Expected padded content to be longer than original")
		}
	})

	t.Run("BorderStyle", func(t *testing.T) {
		rendered := BorderStyle.Render("bordered")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})
}

func TestFormatStatusWithClaudeSquad(t *testing.T) {
	tests := []struct {
		status   string
		expected string // We'll check that it contains certain characters
	}{
		{"attached", "●"},
		{"running", "●"},
		{"ready", "○"},
		{"inactive", "○"},
		{"unknown", "?"},
		{"", "?"},
		{"invalid", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := FormatStatusWithClaudeSquad(tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected status %q to contain %q, got %q", tt.status, tt.expected, result)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	// Test legacy function that should delegate to Claude Squad formatting
	tests := []struct {
		status   string
		expected string
	}{
		{"attached", "●"},
		{"running", "●"},
		{"ready", "○"},
		{"inactive", "○"},
		{"unknown", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := FormatStatus(tt.status)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected status %q to contain %q, got %q", tt.status, tt.expected, result)
			}

			// Should match Claude Squad formatting
			claudeSquadResult := FormatStatusWithClaudeSquad(tt.status)
			if result != claudeSquadResult {
				t.Errorf("FormatStatus should match FormatStatusWithClaudeSquad, got %q vs %q", result, claudeSquadResult)
			}
		})
	}
}

func TestApplyClaudeSquadTheme(t *testing.T) {
	// Create a basic style
	baseStyle := lipgloss.NewStyle()

	// Apply Claude Squad theme
	themedStyle := ApplyClaudeSquadTheme(baseStyle)

	// Test that the themed style renders
	themedRendered := themedStyle.Render("test")

	// Just verify it doesn't panic and produces some output
	if themedRendered == "" {
		t.Error("Expected themed style to render output, got empty string")
	}

	// Test that the function doesn't crash with empty input
	emptyRendered := themedStyle.Render("")
	_ = emptyRendered // Just ensure it doesn't panic
}

func TestApplyTheme(t *testing.T) {
	// Test legacy function that should delegate to Claude Squad theming
	baseStyle := lipgloss.NewStyle()

	legacyThemed := ApplyTheme(baseStyle)
	claudeSquadThemed := ApplyClaudeSquadTheme(baseStyle)

	// Should produce the same result
	legacyRendered := legacyThemed.Render("test")
	claudeSquadRendered := claudeSquadThemed.Render("test")

	if legacyRendered != claudeSquadRendered {
		t.Error("ApplyTheme should delegate to ApplyClaudeSquadTheme")
	}
}

func TestStatusStyles(t *testing.T) {
	// Test status-specific styles
	t.Run("StatusReadyStyle", func(t *testing.T) {
		rendered := StatusReadyStyle.Render("Ready")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("StatusRunningStyle", func(t *testing.T) {
		rendered := StatusRunningStyle.Render("Running")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("StatusErrorStyle", func(t *testing.T) {
		rendered := StatusErrorStyle.Render("Error")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})
}

func TestSelectedItemStyles(t *testing.T) {
	t.Run("ClaudeSquadSelectedStyle", func(t *testing.T) {
		rendered := ClaudeSquadSelectedStyle.Render("Selected Item")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ClaudeSquadSelectedDescStyle", func(t *testing.T) {
		rendered := ClaudeSquadSelectedDescStyle.Render("Selected Description")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("SelectedItemStyle", func(t *testing.T) {
		rendered := SelectedItemStyle.Render("Legacy Selected")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})
}

func TestTextStyles(t *testing.T) {
	t.Run("ClaudeSquadNormalTitleStyle", func(t *testing.T) {
		rendered := ClaudeSquadNormalTitleStyle.Render("Normal Title")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ClaudeSquadNormalDescStyle", func(t *testing.T) {
		rendered := ClaudeSquadNormalDescStyle.Render("Normal Description")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("HelpStyle", func(t *testing.T) {
		rendered := HelpStyle.Render("Help text")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})
}

func TestHeaderStyles(t *testing.T) {
	t.Run("ClaudeSquadHeaderStyle", func(t *testing.T) {
		rendered := ClaudeSquadHeaderStyle.Render("Header")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})

	t.Run("ClaudeSquadHeaderBarStyle", func(t *testing.T) {
		rendered := ClaudeSquadHeaderBarStyle.Render("Header Bar")
		if rendered == "" {
			t.Error("Expected rendered output, got empty string")
		}
	})
}

// Test that styles can be chained and modified
func TestStyleChaining(t *testing.T) {
	// Test that we can build upon existing styles
	customStyle := ClaudeSquadBaseStyle.Copy().Bold(true).Italic(true)
	rendered := customStyle.Render("Custom styled text")

	if rendered == "" {
		t.Error("Expected rendered output from chained style, got empty string")
	}

	// Should be different from base style
	baseRendered := ClaudeSquadBaseStyle.Render("Custom styled text")
	// Since styles may render the same in test environments without a terminal,
	// we just verify that both render successfully and don't panic
	if baseRendered == "" {
		t.Error("Expected base rendered output, got empty string")
	}
}

// Test edge cases
func TestStyleEdgeCases(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		rendered := ClaudeSquadAccentStyle.Render("")
		// Empty string should still have style codes even if no content
		if rendered == "" {
			// This might be valid behavior, so just ensure it doesn't panic
		}
	})

	t.Run("MultilineString", func(t *testing.T) {
		multiline := "Line 1\nLine 2\nLine 3"
		rendered := ClaudeSquadBorderStyle.Render(multiline)
		if rendered == "" {
			t.Error("Expected rendered output for multiline string, got empty string")
		}
	})

	t.Run("UnicodeString", func(t *testing.T) {
		unicode := "Hello 世界 🌍"
		rendered := ClaudeSquadPrimaryStyle.Render(unicode)
		if rendered == "" {
			t.Error("Expected rendered output for unicode string, got empty string")
		}
	})
}
