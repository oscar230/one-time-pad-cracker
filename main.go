package main

import (
	"bufio"
	"encoding/hex"
	"os"
	"strings"

	"github.com/fatih/color"
)

func main() {
	// wordlist := []string{"the", "and", "that", "have", "for", "not", "with", "you", "this", "but", "his", "from", "they", "say", "her", "she", "will", "one", "all", "would", "there", "thier", "what", "hello", "world", "program"}

	f, err := os.Open("cipher")
	check(err)
	var cipherkeys [][]byte
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cipherkeys = append(cipherkeys, hextobyteslice(scanner.Text()))
	}
	defer f.Close()

	f, err = os.Open("wordlist")
	check(err)
	var wordlist []string
	scanner = bufio.NewScanner(f)
	for scanner.Scan() {
		var text = scanner.Text()
		if len(text) > 1 {
			wordlist = append(wordlist, text)
		}
	}
	defer f.Close()

	color.New(color.FgYellow).Printf("This is the \"One-Time Pad cracker\" by Oscar Andersson at KAU.\n")
	color.New(color.FgYellow).Printf("Cipherkeys:\t%d\nWordlist:\t\t%d\n", len(cipherkeys), len(wordlist))
	crack(cipherkeys, wordlist)
}

func crack(cipherkeys [][]byte, wordlist []string) {
	for _, word := range wordlist {
		drag(cipherkeys[0], cipherkeys[1], word, wordlist)
	}
}

func drag(ct1, ct2 []byte, word string, wordlist []string) string {
	wordx := []byte(word)
	ctx := xorbytes(ct1, ct2)
	// color.New(color.FgWhite).Printf("Dragging:\n\tword: %s = %x\n\tct1: %x\n\tct2: %x\n\tctx: %x\n", word, wordx, ct1, ct2, ctx)
	for i := 0; i < len(ctx)-len(wordx)+1; i++ {
		result := xorbytes(ctx[i:i+len(wordx)], wordx)
		// color.New(color.FgWhite).Printf("\t\tDrag %x", ctx[0:i])
		// color.New(color.FgGreen).Printf("%x", ctx[i:i+len(wordx)])
		// if i+len(wordx) < len(ctx) {
		// 	color.New(color.FgWhite).Printf("%x", ctx[i+len(wordx):len(ctx)])
		// }
		// color.New(color.FgWhite).Printf(" at %d => %x %s\n", i, result, result)
		for _, w := range wordlist {
			if strings.Contains(strings.ToLower(string(result)), w) {
				color.New(color.FgHiWhite).Add(color.Bold).Printf("Found")
				color.New(color.FgHiBlue).Add(color.Bold).Printf(" %s ", result)
				color.New(color.FgHiWhite).Add(color.Bold).Printf("matched in wordlist at")
				color.New(color.FgHiGreen).Add(color.Bold).Printf(" %s\n", w)
				return w
			}
		}
	}
	return ""
}

func xorbytes(a, b []byte) (c []byte) {
	if len(a) != len(b) {
		//Not same lenght prepend 0:s
		if len(a) > len(b) {
			b = prependzero(b, len(a)-len(b))
		} else {
			a = prependzero(a, len(b)-len(a))
		}
	}
	for i := 0; i < len(a); i++ {
		c = append(c, a[i]^b[i])
	}
	//color.New(color.FgHiMagenta).Printf("%x ^\n%x =\n%x\n", a, b, c)
	return c
}

func prependzero(bs []byte, amount int) (output []byte) {
	var zero []byte
	for i := 0; i < amount; i++ {
		zero = append(zero, []byte{0}...)
		color.New(color.FgHiBlue).Printf("zero%+v\n", zero)
	}
	output = append(zero, bs...)
	color.New(color.FgHiBlue).Printf("prepend %d zeroes.\n%b\n%b\n", amount, bs, output)
	return output
}

func hextobyteslice(input string) (output []byte) {
	output, err := hex.DecodeString(input)
	check(err)
	return output
}

func check(err error) {
	if err != nil {
		color.New(color.FgRed).Add(color.Bold).Println(err)
	}
}
