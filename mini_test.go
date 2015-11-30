package mini

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_next(t *testing.T) {
	tests := []struct {
		input  *lexer
		output string
	}{
		{
			input:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("foo")))},
			output: "foo",
		},
		{
			input:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("\nfoo")))},
			output: "foo",
		},
		{
			input:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("\tfoo")))},
			output: "foo",
		},
		{
			input:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte(" foo")))},
			output: "foo",
		},
	}

	for _, test := range tests {
		test.input.next()
		assert.Equal(t, test.input.line, test.output, "expected %s but got %s", test.output, test.input.line)
	}
}

func Test_sect(t *testing.T) {

	tests := []struct {
		lexer  *lexer
		input  string
		output string
	}{
		{
			lexer:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo]")))},
			input:  "[foo]",
			output: "[foo]",
		},
		{
			lexer:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo")))},
			input:  "[foo",
			output: "mini: section names must be surrounded by [ and ], as in [section]",
		},
		{
			lexer:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("foo]")))},
			input:  "foo]",
			output: "foo]",
		},
		{
			lexer:  &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("foo")))},
			input:  "foo",
			output: "mini: section names must be surrounded by [ and ], as in [section]",
		},
	}

	for _, test := range tests {
		test.lexer.sections = make(map[string]*section)
		test.lexer.line = test.input
		err := test.lexer.sect()

		if err != nil {
			assert.Equal(t, err.Error(), test.output, "")
		} else {
			section := test.lexer.section
			assert.NotNil(t, section, "should have a section")
		}
	}
}

func Test_cont(t *testing.T) {

	tests := []struct {
		lexer *lexer
		input string
		err   string
		key   string
		value string
	}{
		{
			lexer: &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo]")))},
			input: "foo = bar",
			key:   "foo",
			value: "bar",
		},
		{
			lexer: &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo")))},
			input: "foo bar",
			err:   "mini: configuration format requires an equals between the key and value",
		},
		{
			lexer: &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("foo")))},
			input: "= foo bar",
			err:   "mini: configuration format requires an equals between the key and value",
		},
	}
	for _, test := range tests {
		test.lexer.sections = make(map[string]*section)
		test.lexer.section = &section{"test", make(map[string]interface{})}
		test.lexer.line = test.input
		err := test.lexer.cont()

		if err != nil {
			assert.Equal(t, err.Error(), test.err, "")
		} else {
			section := test.lexer.section
			assert.NotNil(t, section, "should have a section")
			assert.Equal(t, section.name, "test")
			assert.Equal(t, section.values["foo"], "bar")
		}
	}
}

func Test_start(t *testing.T) {

	tests := []struct {
		lexer     *lexer
		sections  map[string]*section
		globals   *section
		isSection bool
	}{
		{
			lexer: &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo]\nbar=baz")))},
			sections: map[string]*section{
				"foo": &section{
					name: "foo",
					values: map[string]interface{}{
						"bar": "baz",
					},
				},
			},
			isSection: true,
		},
		{
			lexer: &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("[foo]\nbar=baz\n[ni]\nkey=value\nfoo=eggs")))},
			sections: map[string]*section{
				"foo": &section{
					name: "foo",
					values: map[string]interface{}{
						"bar": "baz",
					},
				},
				"ni": &section{
					name: "ni",
					values: map[string]interface{}{
						"key": "value",
						"foo": "eggs",
					},
				},
			},
			isSection: true,
		},
		{
			lexer:     &lexer{scanner: bufio.NewScanner(bytes.NewReader([]byte("bar=baz")))},
			globals:   &section{"", map[string]interface{}{"bar": "baz"}},
			isSection: false,
		},
	}

	for _, test := range tests {
		test.lexer.sections = make(map[string]*section)
		test.lexer.globals = &section{"", make(map[string]interface{})}
		err := test.lexer.start()
		if err != nil {
			t.Log(err)
		}
		if test.isSection {
			if !reflect.DeepEqual(test.sections, test.lexer.sections) {
				t.Errorf("expected %#v got %#v", test.sections, test.lexer.sections)
				t.Errorf("expected %#v", test.sections["foo"])
				t.Errorf("expected %#v", test.lexer.sections["foo"])
				t.Fail()
			}
		} else {
			if !reflect.DeepEqual(test.globals, test.lexer.globals) {
				t.Errorf("expected %#v got %#v", test.sections, test.lexer.globals)
			}
		}
	}
}

func Test_getType(t *testing.T) {
	type testStruct struct{}
	tests := []struct {
		input  interface{}
		output reflect.Kind
	}{
		{
			input:  make(map[string]bool),
			output: reflect.Map,
		},
		{
			input:  testStruct{},
			output: reflect.Struct,
		},
		{
			input:  &testStruct{},
			output: reflect.Struct,
		},
	}
	for _, test := range tests {
		itemType := getType(test.input)
		if itemType != test.output {
			t.Errorf("expected %s got %s", test.output, itemType)
			t.Fail()
		}
	}
}

func Test_populateField(t *testing.T) {

	tests := []struct{
		config interface{}
		values map[string]string
	}{
		config: &struct{
			Foo string `key:"foo"`
		},
		values: map[string]string{"foo": "bar"},
	}
}
