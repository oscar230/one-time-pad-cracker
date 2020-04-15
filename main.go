package main

import (
	"bufio"
	"encoding/hex"
	"os"
	"time"

	"github.com/fatih/color"
)

func main() {
	color.New(color.FgYellow).Printf("This is the \"One-Time Pad cracker\" by Oscar Andersson at KAU.\n")

	cipherkeys := loadcipher()
	wordlist := loadwordlist()

	color.New(color.FgYellow).Printf("Cipherkeys: %d\nWordlist: %d\nCombinations: %d\n", len(cipherkeys), len(wordlist), len(cipherkeys)*len(cipherkeys)*len(wordlist))

	crack(cipherkeys, wordlist)
}

func scan(cipherkeys [][]byte, dragged [][]string, wordlist []string) {
	for i, _ := range cipherkeys {
		color.New(color.FgHiCyan).Printf("Cipher %d with %d dragged words.\n", i, len(dragged[i]))
		// for _, draggedword := range dragged[i] {
		// 	color.New(color.FgHiCyan).Printf("\t%s\n", draggedword)
		// }
	}
}

func crack(cipherkeys [][]byte, wordlist []string) {
	var dragged = make([][]string, len(cipherkeys))

	// Drag phase
	color.New(color.FgGreen).Add(color.Bold).Printf("Start multithreaded dragging.")
	timestart := time.Now()
	for i, c1 := range cipherkeys {
		for _, c2 := range cipherkeys {
			for _, word := range wordlist {
				var dragchannel chan []string = make(chan []string)
				go drag(c1, c2, word, dragchannel)
				var dw = <-dragchannel
				dragged[i] = append(dragged[i], dw...)
			}
		}
		color.New(color.FgGreen).Printf("Cipher %d done.\n", i)
	}
	color.New(color.FgGreen).Add(color.Bold).Printf("All ciphers total time: %ds.\n", time.Now().Unix()-timestart.Unix())

	// Analysis phase
	scan(cipherkeys, dragged, wordlist)
}

func drag(ct1, ct2 []byte, word string, dragchannel chan []string) (dragged []string) {
	wordx := []byte(word)
	ctx := xorbytes(ct1, ct2)
	passes := len(ctx) - len(wordx) + 1
	if passes < 0 {
		passes = 0
	}
	// color.New(color.FgWhite).Printf("Cipher: ")
	// color.New(color.FgHiGreen).Printf("%x", ct1)
	// color.New(color.FgWhite).Printf(" & ")
	// color.New(color.FgHiGreen).Printf("%x", ct2)
	// color.New(color.FgWhite).Printf("\twill do ")
	// color.New(color.FgHiGreen).Printf("%d", passes)
	// color.New(color.FgWhite).Printf("\tpasses with word: ")
	// color.New(color.FgHiGreen).Printf("%s\n", word)
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
		wordlist = append(wordlist, scanner.Text())
	}
	defer f.Close()
	return
}
