package mini

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strings"
)

//func Parse(reader io.Reader, config interface{}) error {
//lexer := newLexer(reader)
//err := lexer.start()
//if err != nil && err != io.EOF {
//return err
//}

//populateConfig(lexer, config)

//}

func populateConfig(lexer *lexer, config interface{}) error {
	t := reflect.TypeOf(config)
	for i := 0; i < t.NumField(); i++ {
		switch t.Field(i) {
		case struct:
			err := populateConfig(lexer, config)
		default:
			err := populateValue(lexer, config)
		}

	}

}

func getType(field interface{}) reflect.Kind {
	return reflect.Indirect(reflect.ValueOf(field)).Kind()
}

type section struct {
	name   string
	values map[string]interface{}
}

func newSection(name string) *section {
	return &section{
		name:   name,
		values: make(map[string]interface{}),
	}
}

type lexer struct {
	scanner  *bufio.Scanner
	line     string
	section  *section
	globals  *section
	sections map[string]*section
}

func newLexer(reader io.Reader) *lexer {

	return &lexer{
		scanner:  bufio.NewScanner(reader),
		globals:  newSection(""),
		sections: make(map[string]*section),
	}
}

func (l *lexer) next() error {

	for l.scanner.Scan() {
		line := l.scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		l.line = line
		return nil
	}
	return io.EOF
}

func (l *lexer) sect() error {
	line := l.line

	if !strings.HasSuffix(line, "]") {
		return fmt.Errorf("mini: section names must be surrounded by [ and ], as in [section]")
	}

	name := line[1 : len(line)-1]
	if section, ok := l.sections[name]; !ok {
		l.section = newSection(name)
		l.sections[name] = l.section
	} else {
		l.section = section
	}

	return nil
}

func (l *lexer) cont() error {
	line := l.line
	index := strings.Index(line, "=")
	if index <= 0 {
		return fmt.Errorf("mini: configuration format requires an equals between the key and value")
	}
	key := strings.ToLower(strings.TrimSpace(line[0:index]))
	value := strings.TrimSpace(line[index+1:])
	value = strings.Trim(value, "\"'")

	values := l.globals.values
	if l.section != nil {
		values = l.section.values
	}
	values[key] = value
	return nil
}

func (l *lexer) start() error {

	for {
		err := l.next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch l.line[0] {
		case '[':
			err = l.sect()
		default:
			err = l.cont()

		}
		if err != nil {
			return err
		}
	}
	return l.scanner.Err()
}
