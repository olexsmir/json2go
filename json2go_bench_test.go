package json2go

import (
	"testing"
)

func BenchmarkTransform(b *testing.B) {
	benches := map[string]string{
		"simple_object":    `{"name":"alice","age":30,"active":true}`,
		"flat_object":      `{"id":1,"first_name":"john","last_name":"doe","email":"john@example.com","phone":"+1234567890","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}`,
		"nested_object":    `{"user":{"id":1,"name":"alice","profile":{"bio":"engineer","location":"sf","skills":["go","rust","python"]},"settings":{"theme":"dark","notifications":true}}}`,
		"array_of_objects": `[{"id":1,"name":"alice"},{"id":2,"name":"bob"},{"id":3,"name":"charlie"}]`,
		"mixed_types":      `{"string":"text","number":42,"decimal":3.14,"bool":true,"null_val":null,"array":[1,2,3],"object":{"nested":true}}`,
		"deeply_nested":    `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":{"j":{"value":"deep"}}}}}}}}}}}`,
		"large_array":      `[{"id":1,"val":"a"},{"id":2,"val":"b"},{"id":3,"val":"c"},{"id":4,"val":"d"},{"id":5,"val":"e"},{"id":6,"val":"f"},{"id":7,"val":"g"},{"id":8,"val":"h"},{"id":9,"val":"i"},{"id":10,"val":"j"}]`,
		"unicode_strings":  `{"english":"hello","chinese":"你好","arabic":"مرحبا","emoji":"🚀🔥✨"}`,
	}
	for bname, bjson := range benches {
		b.Run(bname+"_with_tags", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				Transform("Test", bjson, true)
			}
		})

		b.Run(bname+"_no_tags", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				Transform("Test", bjson, false)
			}
		})
	}
}

func BenchmarkPipeline(b *testing.B) {
	json := `{"user":{"id":1,"name":"alice","email":"alice@example.com","profile":{"bio":"engineer","location":"sf"}}}`

	b.Run("lexer_only", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lexer := NewLexer([]byte(json))
			for {
				tok := lexer.Next()
				if tok.Type == EOF {
					break
				}
			}
		}
	})

	b.Run("parser_only", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lexer := NewLexer([]byte(json))
			parser := NewParser(lexer)
			parser.Parse()
		}
	})

	b.Run("transpiler_only", func(b *testing.B) {
		lexer := NewLexer([]byte(json))
		parser := NewParser(lexer)
		v, _ := parser.Parse()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			NewTranspiler().Transpile("Test", v, true)
		}
	})

	b.Run("full_pipeline", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			Transform("Test", json, true)
		}
	})
}
