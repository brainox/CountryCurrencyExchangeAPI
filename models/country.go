package models

import (
	"fmt"
	"math/rand"
	"strings"

	"country-currency-exchange-api/database"
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
	ID              int64      `json:"id"`
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

// get all countries from database
func GetAllCountries() []Country {
	query := `SELECT 
		id, name, capital, region, population,
		currency_code, exchange_rate, estimated_gdp,
		flag_url, last_refreshed_at 
	FROM countries`

	rows, err := database.DB.Query(query)
	if err != nil {
		fmt.Printf("Error querying countries: %v\n", err)
		return nil
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var country Country
		err := rows.Scan(
			&country.ID,
			&country.Name,
			&country.Capital,
			&country.Region,
			&country.Population,
			&country.CurrencyCode,
			&country.ExchangeRate,
			&country.EstimatedGDP,
			&country.Flag,
			&country.LastRefreshedAt,
		)
		if err != nil {
			fmt.Printf("Error scanning country: %v\n", err)
			continue
		}
		countries = append(countries, country)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
		return nil
	}

	return countries
}

// SaveCountries saves multiple countries to the database
func SaveCountries(newCountries []Country) error {
	// Clear existing data
	_, err := database.DB.Exec("DELETE FROM countries")
	if err != nil {
		return fmt.Errorf("error clearing existing countries: %v", err)
	}

	// Prepare the insert statement
	stmt, err := database.DB.Prepare(`
		INSERT INTO countries (
			name, capital, region, population, 
			currency_code, exchange_rate, estimated_gdp, 
			flag_url, last_refreshed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	// Insert each country
	for _, country := range newCountries {
		_, err := stmt.Exec(
			country.Name,
			country.Capital,
			country.Region,
			country.Population,
			country.CurrencyCode,
			country.ExchangeRate,
			country.EstimatedGDP,
			country.Flag,
			country.LastRefreshedAt,
		)
		if err != nil {
			return fmt.Errorf("error inserting country %s: %v", country.Name, err)
		}
	}

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
