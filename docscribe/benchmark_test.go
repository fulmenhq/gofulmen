package docscribe

import (
	"bytes"
	"strings"
	"testing"
)

// BenchmarkInspectDocument benchmarks document inspection
// Target: <1ms for 100KB documents
func BenchmarkInspectDocument(b *testing.B) {
	// Create a ~100KB markdown document
	doc := generate100KBMarkdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := InspectDocument(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseFrontmatter benchmarks frontmatter parsing
// Target: <5ms
func BenchmarkParseFrontmatter(b *testing.B) {
	doc := generateDocWithFrontmatter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ParseFrontmatter(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStripFrontmatter benchmarks frontmatter stripping
func BenchmarkStripFrontmatter(b *testing.B) {
	doc := generateDocWithFrontmatter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StripFrontmatter(doc)
	}
}

// BenchmarkExtractHeaders benchmarks header extraction
// Target: <50ms for 1MB document
func BenchmarkExtractHeaders(b *testing.B) {
	doc := generate1MBMarkdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExtractHeaders(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDetectFormat benchmarks format detection
func BenchmarkDetectFormat(b *testing.B) {
	docs := []struct {
		name    string
		content []byte
	}{
		{"markdown", generateDocWithFrontmatter()},
		{"json", []byte(`{"key": "value", "array": [1, 2, 3]}`)},
		{"yaml", []byte("key: value\nlist:\n  - item1\n  - item2")},
	}

	for _, doc := range docs {
		b.Run(doc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = DetectFormat(doc.content)
			}
		})
	}
}

// BenchmarkSplitDocuments benchmarks document splitting
// Target: <10ms for 10-document stream
func BenchmarkSplitDocuments(b *testing.B) {
	benchmarks := []struct {
		name string
		doc  []byte
	}{
		{"single-doc", generateDocWithFrontmatter()},
		{"yaml-stream-10", generateYAMLStream(10)},
		{"multi-markdown-5", generateMultiMarkdown(5)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := SplitDocuments(bm.doc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGenerateAnchor benchmarks anchor slug generation
func BenchmarkGenerateAnchor(b *testing.B) {
	headers := []string{
		"Simple Header",
		"Header with Special Characters!@#$%",
		"Very Long Header with Many Words and Complex Characters (v2.0)",
	}

	for _, header := range headers {
		b.Run(header, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = generateAnchor(header)
			}
		})
	}
}

// Helper functions to generate test documents

func generate100KBMarkdown() []byte {
	var buf bytes.Buffer
	buf.WriteString("---\ntitle: Test Document\nauthor: Test\n---\n\n")

	// Generate content to reach ~100KB
	for i := 0; i < 500; i++ {
		buf.WriteString("## Section ")
		buf.WriteString(string(rune('A' + (i % 26))))
		buf.WriteString("\n\n")
		buf.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
		buf.WriteString("Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n")
	}

	return buf.Bytes()
}

func generate1MBMarkdown() []byte {
	var buf bytes.Buffer
	buf.WriteString("# Main Document\n\n")

	// Generate content to reach ~1MB
	for i := 0; i < 5000; i++ {
		buf.WriteString("## Section ")
		buf.WriteString(strings.Repeat("=", i%10+1))
		buf.WriteString("\n\n")
		buf.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ")
		buf.WriteString("Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ")
		buf.WriteString("Ut enim ad minim veniam, quis nostrud exercitation ullamco.\n\n")
	}

	return buf.Bytes()
}

func generateDocWithFrontmatter() []byte {
	return []byte(`---
title: "Benchmark Document"
description: "A test document for benchmarking"
author: "Benchmark Suite"
tags: ["test", "benchmark", "performance"]
---

# Main Title

This is a test document used for benchmarking.

## Section 1

Content for section 1.

## Section 2

Content for section 2.

### Subsection 2.1

More content here.
`)
}

func generateYAMLStream(count int) []byte {
	var buf bytes.Buffer

	for i := 0; i < count; i++ {
		if i > 0 {
			buf.WriteString("---\n")
		}
		buf.WriteString("apiVersion: v1\n")
		buf.WriteString("kind: Pod\n")
		buf.WriteString("metadata:\n")
		buf.WriteString("  name: pod-")
		buf.WriteString(string(rune('0' + i)))
		buf.WriteString("\nspec:\n")
		buf.WriteString("  containers:\n")
		buf.WriteString("    - name: nginx\n")
		buf.WriteString("      image: nginx:latest\n")
	}

	return buf.Bytes()
}

func generateMultiMarkdown(count int) []byte {
	var buf bytes.Buffer

	for i := 0; i < count; i++ {
		if i > 0 {
			buf.WriteString("---\n")
		}
		buf.WriteString("---\n")
		buf.WriteString("title: \"Document ")
		buf.WriteString(string(rune('0' + i)))
		buf.WriteString("\"\n")
		buf.WriteString("---\n\n")
		buf.WriteString("# Document ")
		buf.WriteString(string(rune('0' + i)))
		buf.WriteString("\n\n")
		buf.WriteString("Content for document ")
		buf.WriteString(string(rune('0' + i)))
		buf.WriteString(".\n\n")
	}

	return buf.Bytes()
}
