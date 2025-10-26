package input

import "github.com/hajimehoshi/ebiten/v2"

// punctuation → rune map
var punctuationMap = map[ebiten.Key]rune{
    ebiten.KeyComma:       ',',
    ebiten.KeyPeriod:      '.',
    ebiten.KeySlash:       '/',
    ebiten.KeyMinus:       '-',
    ebiten.KeyEqual:       '=',
    ebiten.KeySemicolon:   ';',
    ebiten.KeyQuote:       '\'',
    ebiten.KeyLeftBracket: '[',
    ebiten.KeyRightBracket: ']',
    ebiten.KeyBackslash:   '\\',
}

// shifted punctuation
var shiftPunct = map[rune]rune{
    '1': '!', '2': '@', '3': '#', '4': '$', '5': '%',
    '6': '^', '7': '&', '8': '*', '9': '(', '0': ')',
    '-': '_', '=': '+', '[': '{', ']': '}', ';': ':',
    '\'': '"', ',': '<', '.': '>', '/': '?', '\\': '|',
}

