package models

import (
	"math/rand"
	"strings"
	"fmt"
)

// Temporary struct to match the restcountries.com API response
type RestCountry struct {
	Name       string     `json:"name"`
	Capital    string     `json:"capital"`
	Region     string     `json:"region"`
	Population int64      `json:"population"`
	Currencies []Currency `json:"currencies"`
	Flag       string     `json:"flag"`
}

type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Country struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Capital         string     `json:"capital"`
	Region          string     `json:"region"`
	Population      int64      `json:"population"`
	Currencies      []Currency `json:"-"`
	CurrencyCode    string     `json:"currency_code,omitempty"`
	ExchangeRate    float64    `json:"exchange_rate,omitempty"`
	EstimatedGDP    float64    `json:"estimated_gdp,omitempty"`
	Flag            string     `json:"flag_url"`
	LastRefreshedAt string     `json:"last_refreshed_at"`
}

var countries = []Country{}

// Global variable to track last refresh time
var lastRefreshedAt string

// GetTotalCountries returns the total number of countries
func GetTotalCountries() int {
	return len(countries)
}

// GetLastRefreshedAt returns the last refresh timestamp
func GetLastRefreshedAt() string {
	return lastRefreshedAt
}

// SetLastRefreshedAt updates the last refresh timestamp
func SetLastRefreshedAt(timestamp string) {
	lastRefreshedAt = timestamp
}

// ComputeEstimatedGDP function to compute EstimatedGDP
func (c *Country) ComputeEstimatedGDP() {
	// Only compute if we have valid currency and exchange rate
	if c.CurrencyCode != "" && c.ExchangeRate > 0 {
		randomFactor := float64(rand.Intn(1001) + 1000) // random number between 1000 and 2000
		c.EstimatedGDP = float64(c.Population) * randomFactor / c.ExchangeRate
	} else {
		c.EstimatedGDP = 0
	}
}

// save country to database
func (c Country) Save() {
	// Implementation for saving country to database goes here.
	countries = append(countries, c)
}

// get all countries from database
func GetAllCountries() []Country {
	// Implementation for retrieving all countries from database goes here.
	return countries
}

// SaveCountries saves multiple countries to the database
func SaveCountries(newCountries []Country) error {
	// For now, just replace the in-memory countries slice
	countries = newCountries
	return nil
}

// DeleteCountryByName removes a country from the database by name (case-insensitive)
func DeleteCountryByName(name string) error {
	searchName := strings.ToLower(strings.TrimSpace(name))
	
	for i := range countries {
		countryName := strings.ToLower(strings.TrimSpace(countries[i].Name))
		if countryName == searchName {
			// Remove the country from the slice
			countries = append(countries[:i], countries[i+1:]...)
			return nil
		}
	}
	
	return fmt.Errorf("country not found")
}