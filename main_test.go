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
		name          string
		cfg           config
		contextName   string
		namespaceName string
		want          string
	}{
		{
			"no colors",
			config{separator: "/"},
			"my-cluster", "default",
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
			"my-cluster", "default",
			"#[fg=blue,bg=default]{{.Context}}#[fg=colour250]:#[fg=green]{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"icon prefix",
			config{icon: "⎈ ", separator: "/"},
			"my-cluster", "default",
			"⎈ {{.Context}}/{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"prodFg highlights context when context contains prod",
			config{ctxFg: "blue", nsFg: "green", prodFg: "red", separator: "/"},
			"my-prod-cluster", "default",
			"#[fg=red]{{.Context}}/#[fg=green]{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"prodFg highlights namespace when namespace contains prod",
			config{ctxFg: "blue", nsFg: "green", prodFg: "red", separator: "/"},
			"my-cluster", "prod-ns",
			"#[fg=blue]{{.Context}}/#[fg=red]{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"prodFg highlights both when both contain prod",
			config{ctxFg: "blue", nsFg: "green", prodFg: "red", separator: "/"},
			"my-prod-cluster", "prod-ns",
			"#[fg=red]{{.Context}}/#[fg=red]{{.Namespace}}#[fg=default#,bg=default]",
		},
		{
			"prodFg ignored when neither contains prod",
			config{ctxFg: "blue", nsFg: "green", prodFg: "red", separator: "/"},
			"my-staging-cluster", "default",
			"#[fg=blue]{{.Context}}/#[fg=green]{{.Namespace}}#[fg=default#,bg=default]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.buildFormat(tt.contextName, tt.namespaceName)
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
