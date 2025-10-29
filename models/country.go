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
	CurrencyCode    string     `json:"currency_code"`
	ExchangeRate    float64    `json:"exchange_rate"`
	EstimatedGDP    float64    `json:"estimated_gdp"`
	Flag            string     `json:"flag_url"` // Will be shown as flag_url in JSON response
	LastRefreshedAt string     `json:"last_refreshed_at"`
}

var countries = []Country{}

// function to compute EstimatedGDP
func (c *Country) ComputeEstimatedGDP() {
	// Use this formular: estimated_gdp = population × random(1000–2000) ÷ exchange_rate.
	randomFactor := float64(rand.Intn(1001) + 1000) // random number between 1000 and 2000

	// Ensure we have a valid exchange rate to prevent division by zero
	if c.ExchangeRate <= 0 {
		c.ExchangeRate = 1.0
	}

	c.EstimatedGDP = float64(c.Population) * randomFactor / c.ExchangeRate
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