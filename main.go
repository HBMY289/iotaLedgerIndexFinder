// main
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/kerl"

	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

var wg = sync.WaitGroup{}
var scanner = bufio.NewScanner(os.Stdin)

func main() {
	for {
		settings := getSettings()
		startTime := time.Now()
		matchedAccPage := getMatchingIndex(settings)
		deltaT := time.Now().Sub(startTime)

		if matchedAccPage.acc >= 0 {
			var pageText string
			if cliArgsHas("-p") {
				pageText = fmt.Sprintf(" and PAGE INDEX '%d'", matchedAccPage.page)
			}
			fmt.Printf("\nFound matching address for ACCOUNT INDEX '%d'%v after %v.\n", matchedAccPage.acc, pageText, deltaT.Round(time.Second))
			break
		}
		fmt.Printf("\nCould not find a match after %v testing the first %d addresses of indexes %d to %d.\nCheck target address or retry with larger values for maximum index and addresses per seed.", deltaT.Round(time.Second), settings.addrsPerSeed, settings.accStart, settings.accEnd)
		if !again() {
			return
		}
	}
	fmt.Println("\nPress Enter to close the program.")
	scanner.Scan()
}

func getMatchingIndex(settings settings) accPage {
	input := make(chan accPage)
	result := accPage{-1, -1}
	workers := runtime.GOMAXPROCS(-1)
	fmt.Println("\nstart searching")
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go checkCandidate(input, settings, &result)
	}
L:
	for i := settings.accStart; i <= settings.accEnd; i++ {
		for j := settings.pageStart; j <= settings.pageEnd; j++ {

			if result.acc >= 0 {
				break L
			}
			candidate := accPage{i, j}
			input <- candidate
		}
	}
	close(input)
	wg.Wait()
	fmt.Println("\nstopped searching")
	return result
}

func checkCandidate(input <-chan accPage, settings settings, result *accPage) {
	for candidate := range input {

		if candidate.page == settings.pageStart {
			fmt.Printf("\rchecking index #%d", candidate.acc)
		}
		seed := mnemonicToSeed(settings.mnemonic, candidate.acc, candidate.page)
		addrs := getAddrsOfSeed(seed, settings.addrsPerSeed)

		for _, addr := range addrs {
			for _, targetAddr := range settings.targetAddr {
				if strings.Contains(addr, targetAddr) {
					*result = candidate
					break
				}
			}

		}
	}
	wg.Done()
}

func mnemonicToSeed(mnemonic string, accountIndex, pageIndex int) string {
	seed := bip39.NewSeed(mnemonic, "")
	bip32RootKey, _ := bip32.NewMasterKey(seed)
	bip44PuposeKey, _ := bip32RootKey.NewChildKey(0x8000002C)
	bip44CoinTypeKey, _ := bip44PuposeKey.NewChildKey(0x8000107A)
	bip44AccountKey, _ := bip44CoinTypeKey.NewChildKey(0x80000000 + uint32(accountIndex))
	bip44PageIndexKey, _ := bip44AccountKey.NewChildKey(0x80000000 + uint32(pageIndex))
	privateKey := bip44PageIndexKey.Key
	chainCode := bip44PageIndexKey.ChainCode
	kerlInput := append(privateKey[0:32], chainCode[0:16]...)
	kerlInput = append(kerlInput, privateKey[16:32]...)
	kerlInput = append(kerlInput, chainCode[0:32]...)
	myKerl := kerl.NewKerl()
	myKerl.Hash.Write(kerlInput)
	trytes, _ := myKerl.SqueezeTrytes(243)
	return trytes
}

func getSettings() settings {
	var settings = settings{}
	settings.mnemonic = "wheel mosquito enroll illness stamp vote tomorrow mandate powder armed fortune buffalo rack mirror elder fun paper between cheap present vast unlock detect birth"
	// settings.targetAddr = []string{"SZE9WDWHUUYGOXQRMZWKHFHSQCVU9NROSNFERAJMT9YFIHHRCKRFSDESFWDPCLPMJFFXLXZISLWKBSKTC"} // will not be found
	// settings.targetAddr = []string{"CRICOFALQY9XBDSPOJAID9TMKMUNYWVN99WEUFOTCNBYZCNALGUCDDMQTHYWZVFMNWBYGBBBDUWKJPAFZ"} //addindex #1 accindex #9
	// settings.targetAddr = []string{"JWTWV9KLWZRORTCQGBHEYZFQLZUIGLGJASFDGQOKAVSYIBKOGONQDZZTLM9IYE9GVBTPBSXEWLIDBQYF9"} //addindex #2 accountindex #99
	// settings.targetAddr = []string{"IMRMBIOYEMBT9MHGDHFJXPBXCQEIE9QUCFFMPMQE9YCTKEUZPHVHPGYC9THLGDARGIOZUFBFKYDDZWMPZ"}  // addindex #3 accindex #50 with balance
	settings.targetAddr = []string{"HHQABKTNV9DDDBCPBFVFKGYLXAAUMMSYOTHSQJUWV9JMFYHRXVEPNCRDNOLINQ9RADCTPZDSEBNJZETOD"} // addindex #4 accindex #9 pageindex #9 with balance

	settings.accStart = 0
	settings.accEnd = 1000
	settings.addrsPerSeed = 100

	if cliArgsHas("-s") {
		getAddressesFromSnapshot(&settings)
	}

	if !cliArgsHas("-sm") {
		getMnemonic(&settings)
	}
	if !cliArgsHas("-st") && !cliArgsHas("-s") {
		getTargetAddress(&settings)
	}
	getIntInput(&settings.addrsPerSeed, "Enter number of addresses to test per seed")
	getIntInput(&settings.accStart, "Enter account index start")
	getIntInput(&settings.accEnd, "Enter account index stop")

	if cliArgsHas("-p") {
		settings.pageEnd = 10
		getIntInput(&settings.pageStart, "Enter page index start")
		getIntInput(&settings.pageEnd, "Enter page index stop")
	}

	return settings
}

func cliArgsHas(wantedArg string) bool {
	for _, arg := range os.Args {
		if arg == wantedArg {
			return true
		}
	}
	return false
}

func getTargetAddress(settings *settings) {
	var addr string
	for {
		fmt.Printf("Enter any address belonging to this Ledger account: ")
		scanner.Scan()
		addr = scanner.Text()
		if err := address.ValidAddress(addr); err == nil {
			break
		}
		fmt.Println("\n\nInvalid address entered.")
	}
	settings.targetAddr = []string{addr[0:81]}
}

func getMnemonic(settings *settings) {
	var words [24]string
	i := 1
	for i <= 24 {
		fmt.Printf("Enter mnemonic word #%d: ", i)
		scanner.Scan()
		word := scanner.Text()
		fmt.Println()
		if isValidWord(word) {
			words[i-1] = word
			i++
		} else {
			fmt.Printf("'%v' is not a valid BIP39 word. Try again.\n", word)
		}
	}
	settings.mnemonic = strings.Join(words[:], " ")
}

func getIntInput(value *int, text string) {
	for {
		fmt.Printf(text+" (press Enter for default=%d): ", *value)
		scanner.Scan()
		input := scanner.Text()
		if input == "" {
			return
		}
		if conv, err := strconv.Atoi(input); err == nil {
			*value = conv
			return
		}
		fmt.Println("Invalid input!")
	}

}

func getAddressesFromSnapshot(settings *settings) {
	snapFile := "snapshot.txt"
	data, err := ioutil.ReadFile(snapFile)
	if err != nil {
		fmt.Printf("Could not open snapshot file. Make sure %s is available in the same folder.\n", snapFile)
		fmt.Println(err)
		os.Exit(1)
	}
	var snapshot Snapshot
	json.Unmarshal([]byte(data), &snapshot)
	keys := make([]string, 0, len(snapshot.Balances))
	for key := range snapshot.Balances {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		fmt.Println("Could not read addresses from snapshot file!", err)
		os.Exit(1)
	}
	fmt.Printf("\nSuccessfully read %d addresses from snapshot file.\n", len(keys))
	settings.targetAddr = keys
	return
}

func again() bool {
	fmt.Print("\nDo you want to try again using the same 24 words (y/n)?: ")
	scanner.Scan()
	answer := scanner.Text()
	if answer == "y" {
		fmt.Println()
		return true
	}
	return false
}

func isValidWord(word string) bool {
	_, valid := bip39.GetWordIndex(word)
	return valid
}

func getAddrsOfSeed(seed string, addCount int) []string {
	addrs, _ := address.GenerateAddresses(seed, 0, uint64(addCount)-1, 2, false)
	return addrs
}

type mySeed struct {
	string                  string
	accountIndex, pageIndex int
}

type settings struct {
	mnemonic                                           string
	targetAddr                                         []string
	accStart, accEnd, pageStart, pageEnd, addrsPerSeed int
}

type accPage struct {
	acc, page int
}

type Snapshot struct {
	Balances       map[string]uint64 `json:"balances"`
	MilestoneIndex uint64
	Duration       int `json:"duration"`
}
