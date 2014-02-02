package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-on/queue"
	. "github.com/go-on/queue/q"
)

/*
	This example shows Fallback(), custom errors with Fallback() and logging
*/

func main() {
	codes := []string{"fr_CH", "CH", "fr_BE", "IT", "abc"}

	// our custom error handler
	eh := queue.ErrHandlerFunc(func(err error) error {
		switch err.(type) {
		// stop the queue on InvalidCode
		case InvalidCode:
			return err
			// otherwise continue
		default:
			return nil
		}
	})

	for _, code := range codes {

		fmt.Printf("\n---- CODE %#v\n", code)
		l := &Locale{}

		_, err := Err(eh)(
			l.SetByLanguage, code,
		)(
			l.SetByCountry, code,
		)(
			l.SetDefault,
		).
			LogDebugTo(os.Stdout).
			Fallback()

		// fmt.Printf("\nCode %s ", code)
		if err != nil {
			fmt.Printf("\nError: %s\n", err)
			continue
		}

		fmt.Printf("\n%#v\n", l)
	}
}

var languages = map[string]string{
	"de": "German",
	"fr": "French",
	"en": "English",
}

var countries = map[string]string{
	"DE": "Germany",
	"CH": "Switzerland",
	"US": "USA",
	"FR": "France",
}

var countriesDefaultLanguage = map[string]string{
	"DE": "de",
	"CH": "de",
	"US": "en",
}

var languagesDefaultCountries = map[string]string{
	"en": "US",
	"de": "DE",
	"fr": "FR",
}

type Locale struct {
	Language, Country string
}

type InvalidCode struct {
	msg string
}

func (i InvalidCode) Error() string {
	return fmt.Sprintf(`Wrong code syntax: %#v`, i.msg)
}

var codeRegex = regexp.MustCompile("^([a-z]{2})?_?([A-Z]{2})$")

func (l *Locale) splitCode(code string) (lang, country string, err error) {
	m := codeRegex.FindSubmatch([]byte(code))

	switch len(m) {
	case 0, 1:
		err = InvalidCode{code}
		return
	case 2:
		country = string(m[1])
	case 3:
		lang = string(m[1])
		country = string(m[2])
	}
	return
}

func (l *Locale) SetByCountry(code string) error {
	_, country, err := l.splitCode(code)
	if err != nil {
		return err
	}
	c, has := countries[country]
	if !has {
		return fmt.Errorf("can't find country: %#v", country)
	}

	lang, hasDefault := countriesDefaultLanguage[country]
	if !hasDefault {
		return fmt.Errorf("can't find default language for country: %#v", c)
	}

	l.Country = c
	l.Language = languages[lang]
	return nil
}

func (l *Locale) SetByLanguage(code string) error {
	lang, country, err := l.splitCode(code)
	if err != nil {
		return err
	}

	la, hasLang := languages[lang]
	if !hasLang {
		return fmt.Errorf("can't find language: %#v", lang)
	}

	l.Language = la
	c, has := countries[country]
	if !has {
		country = languagesDefaultCountries[lang]
		c = countries[country]
	}
	l.Country = c

	return nil
}

func (l *Locale) SetDefault() {
	l.SetByLanguage("en_US")
}
