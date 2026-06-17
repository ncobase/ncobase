package service

import "ncobase/plugin/ai/structs"

var actionRegistry = map[string]structs.ActionDescriptor{
	"content.summary": {
		Key:         "content.summary",
		Domain:      "content",
		Name:        "Content summary",
		Description: "Summarize long-form content into a concise editorial summary.",
		OutputType:  "markdown",
	},
	"content.title": {
		Key:         "content.title",
		Domain:      "content",
		Name:        "Title suggestions",
		Description: "Generate clear and searchable content title candidates.",
		OutputType:  "json",
	},
	"content.seo": {
		Key:         "content.seo",
		Domain:      "content",
		Name:        "SEO metadata",
		Description: "Draft SEO title, description, keywords, and social preview text.",
		OutputType:  "json",
	},
	"content.tags": {
		Key:         "content.tags",
		Domain:      "content",
		Name:        "Tag suggestions",
		Description: "Suggest normalized taxonomy tags from content.",
		OutputType:  "json",
	},
	"content.translation": {
		Key:         "content.translation",
		Domain:      "content",
		Name:        "Translation draft",
		Description: "Translate content while preserving structure, terminology, and tone.",
		OutputType:  "markdown",
	},
	"content.review": {
		Key:         "content.review",
		Domain:      "content",
		Name:        "Editorial review",
		Description: "Review clarity, risk, completeness, and publishing readiness.",
		OutputType:  "json",
	},
	"resource.summary": {
		Key:         "resource.summary",
		Domain:      "resource",
		Name:        "Resource summary",
		Description: "Summarize extracted resource text or media metadata.",
		OutputType:  "markdown",
	},
	"resource.description": {
		Key:         "resource.description",
		Domain:      "resource",
		Name:        "Media description",
		Description: "Create accessible descriptions for resource or media context.",
		OutputType:  "markdown",
	},
	"resource.tags": {
		Key:         "resource.tags",
		Domain:      "resource",
		Name:        "Resource tags",
		Description: "Suggest searchable resource tags and category candidates.",
		OutputType:  "json",
	},
	"builder.schema": {
		Key:         "builder.schema",
		Domain:      "builder",
		Name:        "Schema draft",
		Description: "Draft entity fields, validation, API, permission, and migration notes.",
		OutputType:  "json",
	},
	"builder.form": {
		Key:         "builder.form",
		Domain:      "builder",
		Name:        "Form design",
		Description: "Draft form layout, states, validation, and UX behavior.",
		OutputType:  "json",
	},
	"builder.api": {
		Key:         "builder.api",
		Domain:      "builder",
		Name:        "API design",
		Description: "Draft REST API contract, DTOs, permission checks, and failure states.",
		OutputType:  "json",
	},
	"builder.menu": {
		Key:         "builder.menu",
		Domain:      "builder",
		Name:        "Menu design",
		Description: "Draft navigation, route guards, labels, and menu seed requirements.",
		OutputType:  "json",
	},
	"builder.tests": {
		Key:         "builder.tests",
		Domain:      "builder",
		Name:        "Test checklist",
		Description: "Draft backend, frontend, permission, and integration test coverage.",
		OutputType:  "json",
	},
	"system.summary": {
		Key:         "system.summary",
		Domain:      "system",
		Name:        "System analysis",
		Description: "Summarize system information and operational signals without mutation.",
		OutputType:  "markdown",
	},
	"payment.summary": {
		Key:         "payment.summary",
		Domain:      "payment",
		Name:        "Payment analysis",
		Description: "Analyze payment context, status, risk, and next-step checklist without mutation.",
		OutputType:  "json",
	},
	"realtime.summary": {
		Key:         "realtime.summary",
		Domain:      "realtime",
		Name:        "Realtime analysis",
		Description: "Analyze realtime events, notification state, and operational signals without mutation.",
		OutputType:  "json",
	},
}

func defaultActionKeys() []string {
	keys := make([]string, 0, len(actionRegistry))
	for key := range actionRegistry {
		keys = append(keys, key)
	}
	return keys
}
