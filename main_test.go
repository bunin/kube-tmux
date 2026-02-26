package main

import "testing"

func TestTmuxColor(t *testing.T) {
	tests := []struct {
		name string
		fg   string
		bg   string
		want string
	}{
		{"both empty", "", "", ""},
		{"fg only", "blue", "", "#[fg=blue]"},
		{"bg only", "", "default", "#[bg=default]"},
		{"both set", "blue", "default", "#[fg=blue,bg=default]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tmuxColor(tt.fg, tt.bg)
			if got != tt.want {
				t.Errorf("tmuxColor(%q, %q) = %q, want %q", tt.fg, tt.bg, got, tt.want)
			}
		})
	}
}

func TestConfigBuildFormat(t *testing.T) {
	tests := []struct {
		name string
		cfg  config
		want string
	}{
		{
			"no colors",
			config{separator: "/"},
			"{{.Context}}/{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"all colors",
			config{
				ctxFg: "blue", ctxBg: "default",
				sepFg: "colour250", sepBg: "",
				nsFg: "green", nsBg: "",
				separator: ":",
			},
			"#[fg=blue,bg=default]{{.Context}}#[fg=colour250]:#[fg=green]{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"ctx colors only",
			config{ctxFg: "red", separator: "/"},
			"#[fg=red]{{.Context}}/{{.Namespace}}#[fg=default#,bg=default]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.buildFormat()
			if got != tt.want {
				t.Errorf("buildFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadKubeContext(t *testing.T) {
	t.Run("returns error when kubeconfig is invalid", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "/nonexistent/path")
		_, err := loadKubeContext()
		if err == nil {
			t.Error("expected error for nonexistent kubeconfig, got nil")
		}
	})
}
