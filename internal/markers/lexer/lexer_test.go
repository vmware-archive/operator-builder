// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package lexer_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu-labs/operator-builder/internal/markers/lexer"
)

func GetTestLexer(buf string) *lexer.Lexer {
	return lexer.NewLexer(bytes.NewBufferString(buf))
}

func TestLexer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []lexer.Lexeme
		focus    bool // if true, run only tests with focus set to true
	}{
		{
			name:  "marker start",
			input: "+test:flag",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "test"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "flag"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "invalid marker start",
			input: "++",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "math operation",
			input: "2+2=4",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with no scope",
			input: "+hello",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeWarning, Value: `marker without scope found at position: {line:1 column:7}, following "+hello"`},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with scope",
			input: "+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker flag with two scopes",
			input: "+hello:new:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "new"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with no scope",
			input: "+planet=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeWarning, Value: `marker without scope found at position: {line:1 column:8}, following "+planet"`},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with scope",
			input: "+galaxy:planet=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "planet"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker arg with two scopes",
			input: "+galaxy:planet:name=earth",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two args",
			input: "+planet:name=earth,solar-system=milky-way",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two scopes and two args",
			input: "+galaxy:planet:name=earth,solar-system=milky-way",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with two scopes and two args one of which is a flag",
			input: "+galaxy:planet:name=earth,current-location",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArg, Value: "current-location"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with single quoted string arg",
			input: "+galaxy:name=milkyway,description='our home system'",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with double quoted string arg",
			input: `+galaxy:name=milkyway,description="our home system"`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker with literal quoted string arg",
			input: "+galaxy:name=milkyway,description=`our home system`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name: "marker with literal quoted multi-line string arg",
			input: `+galaxy:name=milkyway,description=` + "`" + `our home system
			this is where planet earth is located` + "`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system\n\t\t\tthis is where planet earth is located"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name: "marker with literal quoted multi-line string arg in yaml comment",
			input: `# +galaxy:name=milkyway,description=` + "`" + `our home system
			#this is where planet earth is located` + "`",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "galaxy"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "milkyway"},
				{Type: lexer.LexemeArg, Value: "description"},
				{Type: lexer.LexemeStringLiteral, Value: "our home system\nthis is where planet earth is located"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in go comment no space",
			input: "//+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "//"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in go comment with white space",
			input: "//     +hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "//"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in yaml comment no space",
			input: "#+hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "marker in yaml comment with white space",
			input: "#     +hello:world",
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "hello"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "world"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},

		{
			name: "marker with two args in context",
			input: `#+planet:name=earth,solar-system=milky-way
			plant: earth
			`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "planet"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "name"},
				{Type: lexer.LexemeStringLiteral, Value: "earth"},
				{Type: lexer.LexemeArg, Value: "solar-system"},
				{Type: lexer.LexemeStringLiteral, Value: "milky-way"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "fun with rich",
			input: `#+beetle-:dung:mature=0`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeComment, Value: "#"},
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "beetle-"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeScope, Value: "dung"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "mature"},
				{Type: lexer.LexemeIntegerLiteral, Value: "0"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "simple string slice",
			input: `+operator-builder:slice={"hello","world"}`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeStringLiteral, Value: "hello"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeStringLiteral, Value: "world"},
				{Type: lexer.LexemeSliceEnd, Value: "}"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "simple int slice",
			input: `+operator-builder:slice={1,2}`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeIntegerLiteral, Value: "1"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeIntegerLiteral, Value: "2"},
				{Type: lexer.LexemeSliceEnd, Value: "}"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "simple bool slice",
			input: `+operator-builder:slice={true,false}`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeBoolLiteral, Value: "true"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeBoolLiteral, Value: "false"},
				{Type: lexer.LexemeSliceEnd, Value: "}"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "simple mixed slice",
			input: `+operator-builder:slice={"Hello",1,false}`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeStringLiteral, Value: "Hello"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeIntegerLiteral, Value: "1"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeBoolLiteral, Value: "false"},
				{Type: lexer.LexemeSliceEnd, Value: "}"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "empty slice",
			input: `+operator-builder:slice={}`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeSliceEnd, Value: "}"},
				{Type: lexer.LexemeMarkerEnd, Value: "\n"},
				{Type: lexer.LexemeEOF, Value: ""},
			},
		},
		{
			name:  "incomplete slice",
			input: `+operator-builder:slice={"hello","world"`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeStringLiteral, Value: "hello"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeStringLiteral, Value: "world"},
				{Type: lexer.LexemeError, Value: "malformed slice:  at position: {line:1 column:41}, following \"world\""},
			},
		},
		{
			name:  "incomplete slice in args",
			input: `+operator-builder:slice={"hello","world",test=newArg`,
			expected: []lexer.Lexeme{
				{Type: lexer.LexemeMarkerStart, Value: "+"},
				{Type: lexer.LexemeScope, Value: "operator-builder"},
				{Type: lexer.LexemeSeparator, Value: ":"},
				{Type: lexer.LexemeArg, Value: "slice"},
				{Type: lexer.LexemeSliceBegin, Value: "{"},
				{Type: lexer.LexemeStringLiteral, Value: "hello"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeStringLiteral, Value: "world"},
				{Type: lexer.LexemeSliceDelimiter, Value: ","},
				{Type: lexer.LexemeStringLiteral, Value: "test"},
				{Type: lexer.LexemeError, Value: "malformed slice:  at position: {line:1 column:46}, following \"test\""},
			},
		},
	}

	focused := false

	for _, tt := range tests {
		if tt.focus {
			focused = true

			break
		}
	}

	for _, tt := range tests {
		tt := tt
		if focused && !tt.focus {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := GetTestLexer(tt.input)
			go l.Run()
			actual := []lexer.Lexeme{}
			for {
				lexeme := l.NextLexeme()
				testLexeme := lexer.Lexeme{
					Type:  lexeme.Type,
					Value: lexeme.Value,
				}

				actual = append(actual, testLexeme)
				if lexeme.Type == lexer.LexemeEOF || lexeme.Type == lexer.LexemeError {
					break
				}
			}
			require.Equal(t, tt.expected, actual)
		})
	}

	if focused {
		t.Fatalf("testcase(s) still focused")
	}
}
