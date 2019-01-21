package main

import "github.com/pkg/errors"

type lex struct {
	s string
}

func ConvertType(dbus string) (string, error) {
	l := &lex{s: dbus}
	s, err := l.next()
	if err != nil {
		return "", errors.Wrapf(err, "remaining string: %s, starting string: %s", l.s, dbus)
	}
	return s, nil
}

func (l *lex) next() (string, error) {
	t, ok := typeMap[l.s[0:1]]
	if ok {
		l.s = l.s[1:]
		return t, nil
	}
	switch l.s[0] {
	case 'a':
		if l.s[1] == '{' {
			l.s = l.s[2:]
			return l.nextMap()
		}
		l.s = l.s[1:]
		return l.nextArray()
	case '(':
		l.s = l.s[1:]
		return l.nextStruct()
	default:
		return "", errors.New("unknown char: " + l.s[0:1])
	}
}

func (l *lex) nextMap() (string, error) {
	s := "map["
	n, err := l.next()
	if err != nil {
		return "", err
	}
	s += n
	n, err = l.next()
	if err != nil {
		return "", err
	}
	s += n
	if l.s[0] != '}' {
		return "", errors.New("map can only have 2 elements")
	}
	l.s = l.s[1:]
	return n, nil
}

func (l *lex) nextArray() (string, error) {
	n, err := l.next()
	if err != nil {
		return "", err
	}
	return "[]" + n, nil
}

func (l *lex) nextStruct() (string, error) {
	out := "struct{"

	for name := 'A'; l.s[0] != ')'; name++ {
		t, err := l.next()
		if err != nil {
			return "", err
		}
		out += string(name)
		out += " "
		out += t
		out += ";"
	}
	out += "}"
	l.s = l.s[1:]
	return out, nil
}

var typeMap = map[string]string{
	"b": "boolean",
	"y": "byte",
	"n": "int16",
	"q": "uint16",
	"i": "int32",
	"u": "uint32",
	"x": "int64",
	"t": "uint64",
	"f": "float64",
	"s": "string",
	"o": "dbus.ObjectPath",
	"g": "signature",
	"v": "dbus.Variant",
}
