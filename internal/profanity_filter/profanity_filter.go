package profanityfilter

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	goaway "github.com/TwiN/go-away"
)

//Package for setting up profanity filtering, since the data in this database is meant to be streamed to twitch.

/*
* By default, it will filter all words from the English list on this repository:
* https://github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words
*
* However, it will ignore words written in internal/profanity_filter/profanity_whitelist.txt, new-line separated
 */

const profanityListURL = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/refs/heads/master/en"
const whiteListPath = "./internal/profanity_filter/profanity_whitelist.txt"

func InitProfanityFilter() error {
	resp, err := http.Get(profanityListURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching profanity list: unexpected status %d", resp.StatusCode)
	}

	//Get whitelist values
	whiteList, err := loadWhitelist(whiteListPath)
	if err != nil {
		return fmt.Errorf("unable to load white list: %s", err.Error())
	}

	var customList []string
	scanner := bufio.NewScanner(resp.Body)
	
	//Read line by line
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue
		}

		if whiteList[word] {
			continue
		}

		customList = append(customList, word)
	}

	//Override the default go-away profanity map
	goaway.DefaultProfanities = customList
	return scanner.Err()
}

//Loads whitelist
func loadWhitelist(path string) (map[string]bool, error) {
	whiteList := make(map[string]bool)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return whiteList, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if word != "" {
			whiteList[word] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return whiteList, nil
}