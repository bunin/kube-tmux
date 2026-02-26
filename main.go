// SPDX-FileCopyrightText: Copyright 2021 The go-tmux Authors
// SPDX-License-Identifier: BSD-3-Clause

// Command kube-tmux prints Kubernetes context and namespace to tmux status line.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultSeparator       = "/"
	defaultContextFormat   = "{{.Context}}"
	defaultNamespaceFormat = "{{.Namespace}}"
)

type kubeContext struct {
	Context   string
	Namespace string
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.ctxFg, "ctxFg", "", "Context foreground colour")
	flag.StringVar(&cfg.ctxBg, "ctxBg", "", "Context background colour")
	flag.StringVar(&cfg.sepFg, "sepFg", "", "Separator foreground colour")
	flag.StringVar(&cfg.sepBg, "sepBg", "", "Separator background colour")
	flag.StringVar(&cfg.nsFg, "nsFg", "", "Namespace foreground colour")
	flag.StringVar(&cfg.nsBg, "nsBg", "", "Namespace background colour")
	flag.StringVar(&cfg.separator, "separator", defaultSeparator, "Separator of Context and Namespace")
	flag.StringVar(&cfg.icon, "icon", "⎈ ", "Icon prefix before context")
	flag.StringVar(&cfg.prodFg, "prodFg", "", "Namespace foreground colour when context contains 'prod'")
	flag.Parse()
	return cfg
}

func main() {
	cfg := parseFlags()

	kctx, err := loadKubeContext()
	if err != nil {
		fmt.Fprintf(os.Stdout, "[ERROR] %v\n", err)
		return
	}

	if kctx.Namespace == "" {
		if err := printContext(kctx, defaultContextFormat); err != nil {
			fmt.Fprintln(os.Stdout, "[ERROR] could not print kube context")
		}
		return
	}

	var format string
	if flag.NArg() >= 1 {
		format = flag.Arg(0)
	} else {
		format = cfg.buildFormat(kctx.Context, kctx.Namespace)
	}

	if err := printContext(kctx, format); err != nil {
		fmt.Fprintf(os.Stdout, "[ERROR] could not print kube context: %v\n", err)
	}
}

type config struct {
	ctxFg     string
	ctxBg     string
	sepFg     string
	sepBg     string
	nsFg      string
	nsBg      string
	separator string
	icon      string
	prodFg    string
}

func (c config) buildFormat(contextName, namespaceName string) string {
	ctxFg := c.ctxFg
	if c.prodFg != "" && strings.Contains(strings.ToLower(contextName), "prod") {
		ctxFg = c.prodFg
	}
	nsFg := c.nsFg
	if c.prodFg != "" && strings.Contains(strings.ToLower(namespaceName), "prod") {
		nsFg = c.prodFg
	}

	format := c.icon +
		tmuxColor(ctxFg, c.ctxBg) + defaultContextFormat +
		tmuxColor(c.sepFg, c.sepBg) + c.separator +
		tmuxColor(nsFg, c.nsBg) + defaultNamespaceFormat +
		"#[fg=default#,bg=default]"
	return format
}

func loadKubeContext() (kubeContext, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return kubeContext{}, fmt.Errorf("could not get kubeconfig: %w", err)
	}
	if len(rawConfig.Contexts) == 0 {
		return kubeContext{}, fmt.Errorf("kubeconfig is empty")
	}

	curCtx := rawConfig.CurrentContext
	if curCtx == "" {
		return kubeContext{Context: "empty"}, nil
	}

	kctx := kubeContext{Context: curCtx}
	if ns := rawConfig.Contexts[curCtx].Namespace; ns != "" {
		kctx.Namespace = ns
	} else {
		kctx.Namespace = corev1.NamespaceDefault
	}
	return kctx, nil
}

func tmuxColor(fg, bg string) string {
	if fg == "" && bg == "" {
		return ""
	}
	s := "#["
	if fg != "" {
		s += "fg=" + fg
	}
	if fg != "" && bg != "" {
		s += ","
	}
	if bg != "" {
		s += "bg=" + bg
	}
	s += "]"
	return s
}

func printContext(kctx kubeContext, format string) error {
	return template.Must(template.New("kube-tmux").Parse(format)).Execute(os.Stdout, kctx)
}
