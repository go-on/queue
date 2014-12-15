package main

import (
	"fmt"
	"regexp"

	"gopkg.in/go-on/queue.v2"
	. "gopkg.in/go-on/queue.v2/q"
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

		err := Err(eh)(
			Get, &code,
		)(
			l.Set, Fallback(
				Q(SetByLanguage, V),
				Q(SetByCountry, V),
				Q(SetDefault),
			),
		).Run()

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

func (l *Locale) Set(country, language string) {
	l.Language = language
	l.Country = country
}

type InvalidCode struct {
	msg string
}

func (i InvalidCode) Error() string {
	return fmt.Sprintf(`Wrong code syntax: %#v`, i.msg)
}

var codeRegex = regexp.MustCompile("^([a-z]{2})?_?([A-Z]{2})$")

func splitCode(code string) (lang, country string, err error) {
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

func getCountry(country string) (c string, err error) {
	c, has := countries[country]
	if !has {
		err = fmt.Errorf("can't find count: %#v", country)
		return
	}
	return
}

func SetByCountry(code string) (country string, language string, err error) {
	// fmt.Printf("SetByCountry: %#v\n", code)
	_, count, err := splitCode(code)
	if err != nil {
		return
	}

	c, err2 := getCountry(count)
	if err2 != nil {
		err = err2
		return
	}

	lang, hasDefault := countriesDefaultLanguage[count]
	if !hasDefault {
		err = fmt.Errorf("can't find default language for count: %#v", c)
		return
	}

	country = c
	language = languages[lang]
	return
}

func SetByLanguage(code string) (country string, language string, err error) {
	// fmt.Printf("SetByLanguage: %#v\n", code)
	lang, count, err := splitCode(code)
	if err != nil {
		return
	}

	la, hasLang := languages[lang]
	if !hasLang {
		err = fmt.Errorf("can't find language: %#v", lang)
		return
	}

	language = la
	c, has := countries[count]
	if !has {
		count = languagesDefaultCountries[lang]
		c = countries[count]
	}
	country = c

	return
}

func SetDefault() (country string, language string, err error) {
	// fmt.Println("SetDefault")
	return SetByLanguage("en_US")
}
