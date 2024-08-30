package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type TokenType int

const (
	TokenJSONDelimStart TokenType = iota // json.Delim: {
	TokenKey                             // A string constant.
	TokenValue                           // A string constant.

	// no need DelimEnd because we skipped it
)

func getTokenType(i int) TokenType {
	switch {
	case i == 0:
		return TokenJSONDelimStart
	case i%2 == 1:
		return TokenKey
	case i%2 == 0:
		return TokenValue
	default:
		panic(fmt.Sprintf("unexpected index %d", i))
	}
}

func mustString(token json.Token) string {
	if _, ok := token.(string); !ok {
		panic(fmt.Sprintf("unexpected token type: %T, %v", token, token))
	}
	return token.(string)
}
func handleToken(tokenType TokenType, token json.Token, currPair *KeyValuePair, appendKVPair func(KeyValuePair)) {
	switch tokenType {
	case TokenJSONDelimStart:
		return // do nothing
	case TokenKey:
		currPair.Key = mustString(token)
	case TokenValue:
		currPair.Value = mustString(token)
		appendKVPair(*currPair)
		currPair = &KeyValuePair{}
	default:
		panic(fmt.Sprintf("unexpected token type %v, %v", tokenType, token))
	}
}

func GetKeyValuePairs(jsonStream io.Reader) []KeyValuePair {
	var keyValuePairs []KeyValuePair
	currPair := &KeyValuePair{}

	appendKVPair := func(kvp KeyValuePair) {
		keyValuePairs = append(keyValuePairs, kvp)
	}

	// Since this is a translation file, [key, value] must be string string.
	// So the token order is
	//  json.Delim: {
	//  string: key1
	//  string: value1
	//  string: key2
	//  string: value2
	//  ...
	//  string: keyN
	//  string: valueN
	//  json.Delim: }
	//
	//  If above assumption does not hold, below program will break

	dec := json.NewDecoder(jsonStream)
	i := 0
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("token error: %w", err))
		}

		tokenType := getTokenType(i)

		if !dec.More() {
			// last value reached, only token left is json.Delim: }, skip
			handleToken(tokenType, t, currPair, appendKVPair)
			break
		}

		handleToken(tokenType, t, currPair, appendKVPair)
		i++
	}

	return keyValuePairs
}
