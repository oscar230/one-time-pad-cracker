package main

import (
	"bufio"
	"encoding/hex"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {
	color.New(color.FgYellow).Add(color.Bold).Add(color.Underline).Printf("This is the \"One-Time Pad cracker\" by Oscar Andersson at KAU.\n")

	cipherkeys := loadcipher()
	wordlist := loadwordlist()

	color.New(color.FgYellow).Printf("Cipherkeys: %d\nWordlist: %d\nCombinations: %d\n", len(cipherkeys), len(wordlist), len(cipherkeys)*len(cipherkeys)*len(wordlist)-len(cipherkeys))

	crack(cipherkeys, wordlist)
}

func crack(cipherkeys [][]byte, wordlist []string) {
	var dragged = make([][]string, len(cipherkeys))

	// Drag phase
	color.New(color.FgGreen).Add(color.Bold).Add(color.Underline).Printf("Start multithreaded dragging.\n")
	timestart := time.Now()
	for i, c1 := range cipherkeys {
		for j, c2 := range cipherkeys {
			if i != j {
				for _, word := range wordlist {
					var dragchannel chan []string = make(chan []string)
					go drag(c1, c2, word, dragchannel)
					var dw = <-dragchannel
					dragged[i] = append(dragged[i], dw...)
				}
			}
		}
		color.New(color.FgGreen).Printf("-> Cipher %d done.\n", i)
	}
	color.New(color.FgGreen).Add(color.Bold).Printf("All ciphers total time: %ds.\n", time.Now().Unix()-timestart.Unix())

	// Analysis phase
	scan(cipherkeys, dragged, wordlist)
}

func validate(drag, word string) bool {
	drag = strings.ToLower(drag)
	word = strings.ToLower(word)

	// Speical chars
	dragok, err := regexp.MatchString("(^[a-zA-Z]+$)", drag)
	check(err)
	if !dragok {
		return false
	}

	if strings.Contains(drag, word) {
		return true
	}

	return false
}

func checkdragged(draglist []string, wordlist []string) (output []string) {
	for _, dw := range draglist {
		for _, word := range wordlist {
			if validate(dw, word) {
				output = append(output, word)
				// color.New(color.FgHiBlack).Add(color.Underline).Printf("\t%s", dw)
				// color.New(color.FgCyan).Printf(" => ")
				// color.New(color.FgHiBlack).Add(color.Underline).Printf("%s\n", word)
			}
		}
	}
	return
}

func scan(cipherkeys [][]byte, dragged [][]string, wordlist []string) {
	var words [][]string = make([][]string, len(cipherkeys))
	var limit []int = make([]int, len(cipherkeys))
	color.New(color.FgHiCyan).Add(color.Bold).Add(color.Underline).Printf("Scanning of dragged words.\n")
	for i, c := range cipherkeys {
		limit[i] = len(c)
		color.New(color.FgCyan).Printf("-> Cipher %d of length %d with %d dragged words.\n", i, len(c), len(dragged[i]))
		color.New(color.FgHiBlack).Add(color.Underline).Printf("\t%x\n", c)
		words[i] = checkdragged(dragged[i], wordlist)
		color.New(color.FgHiBlack).Printf("\t%d words are valid.\n", len(words[i]))
	}
	color.New(color.FgCyan).Add(color.Bold).Printf("Done scanning.\n")
	color.New(color.FgHiBlue).Add(color.Bold).Add(color.Underline).Printf("Testing possible sentences.\n")
	var sentences [][]string = make([][]string, len(cipherkeys))
	for i, c := range cipherkeys {
		sentences[i] = findsentence(words[i], limit[i])
		color.New(color.FgBlue).Printf("Found %d sentences for %x\n", len(sentences[i]), c)
		for _, s := range sentences[i] {
			color.New(color.FgBlue).Printf("\t%s\n", s)
		}
	}
	var found int
	for _, s := range sentences {
		found += len(s)
	}
	color.New(color.FgBlue).Add(color.Bold).Printf("Done testing possible sentences. Found a total of %d\n", found)
}

func findsentence(words []string, length int) (output []string) {
	for _, sentence1 := range buildsentence("", words, length) {
		var sentence = sentence1[1:]
		if len(sentence) == length {
			output = append(output, sentence)
		}
	}
	return
}

func buildsentence(start string, words []string, length int) (output []string) {
	if length > 0 {
		for i, word := range words {
			// Remaining words
			var remaining = words[i+1:]
			if i > 0 {
				remaining = append(remaining, words[0:i]...)
			}

			// Add sentence with only the word
			output = append(output, start+" "+word)

			// Add sentence with the word and reitterate
			if len(remaining) > 0 {
				var lastout = output[len(output)-1]
				for _, sentence := range buildsentence(lastout, remaining, length-len(lastout)) {
					output = append(output, sentence)
				}
			}
		}
	}
	return
}

func drag(ct1, ct2 []byte, word string, dragchannel chan []string) (dragged []string) {
	wordx := []byte(word)
	ctx := xorbytes(ct1, ct2)
	passes := len(ctx) - len(wordx) + 1

	for i := 0; i < passes; i++ {
		result := xorbytes(ctx[i:i+len(wordx)], wordx)
		dragged = append(dragged, string(result))
	}

	dragchannel <- dragged
	return
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
	return c
}

func prependzero(bs []byte, amount int) (output []byte) {
	var zero []byte
	for i := 0; i < amount; i++ {
		zero = append(zero, []byte{0}...)
	}
	output = append(zero, bs...)
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

func loadcipher() (cipherkeys [][]byte) {
	f, err := os.Open("cipher")
	check(err)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cipherkeys = append(cipherkeys, hextobyteslice(scanner.Text()))
	}
	defer f.Close()
	return
}

func loadwordlist() (wordlist []string) {
	f, err := os.Open("wordlist")
	check(err)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if len(scanner.Text()) > 1 {
			wordlist = append(wordlist, scanner.Text())
		}
	}
	defer f.Close()
	return
}
