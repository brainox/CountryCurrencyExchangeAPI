package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"country-currency-exchange-api/helpers"
	"country-currency-exchange-api/models"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	os := os.Getenv("PORT")

	router.GET("/status", getStatus)
	router.POST("/countries/refresh", refreshCountryData)
	router.GET("/countries", getAllCountries)
	router.GET("/countries/:name", getCountryByName)
	router.DELETE("/countries/:name", deleteCountryByName)
	router.GET("/countries/image", getSummaryImage)

	if os != "" {
		router.Run(":" + os)
		return
	}
	router.Run(":8080")
}

func refreshCountryData(context *gin.Context) {
	err := fetchCountryData()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Country data refreshed successfully"})
}

// Fetch country data from: https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies
func fetchCountryData() error {
	// Fetch country data
	resp, err := http.Get("https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies")
	if err != nil {
		return fmt.Errorf("error fetching country data: %v", err)
	}
	defer resp.Body.Close()

	var restCountries []models.RestCountry
	if err := json.NewDecoder(resp.Body).Decode(&restCountries); err != nil {
		return fmt.Errorf("error decoding country data: %v", err)
	}

	// Convert restCountries to our Country model
	countries := make([]models.Country, len(restCountries))
	for i, rc := range restCountries {
		countries[i] = models.Country{
			Name:       rc.Name,
			Capital:    rc.Capital,
			Region:     rc.Region,
			Population: rc.Population,
			Currencies: rc.Currencies,
			Flag:       rc.Flag,
		}
	}

	// Fetch exchange rates
	exchangeResp, err := http.Get("https://open.er-api.com/v6/latest/USD")
	if err != nil {
		return fmt.Errorf("error fetching exchange rates: %v", err)
	}
	defer exchangeResp.Body.Close()

	var exchangeData struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(exchangeResp.Body).Decode(&exchangeData); err != nil {
		return fmt.Errorf("error decoding exchange rates: %v", err)
	}

	// Process each country
	for i := range countries {
		// Increment the country ID using index + 1
		countries[i].ID = uint(i + 1)

		// Handle currency code
		if len(countries[i].Currencies) > 0 {
			// Get the first currency code
			countries[i].CurrencyCode = countries[i].Currencies[0].Code

			// Try to get exchange rate from the API
			if rate, exists := exchangeData.Rates[countries[i].CurrencyCode]; exists && rate > 0 {
				countries[i].ExchangeRate = rate

				// Generate random GDP multiplier between 1000 and 2000
				multiplier := 1000.0 + rand.Float64()*1000.0

				// Calculate estimated GDP
				countries[i].EstimatedGDP = float64(countries[i].Population) * multiplier / countries[i].ExchangeRate
			} else {
				// Currency code not found in exchange rates API
				countries[i].ExchangeRate = 0    // Will be stored as null
				countries[i].EstimatedGDP = 0    // Will be stored as null
			}
		} else {
			// No currencies available for this country
			countries[i].CurrencyCode = ""       // Will be stored as null
			countries[i].ExchangeRate = 0        // Will be stored as null
			countries[i].EstimatedGDP = 0        // Will be stored as 0
		}

		// Set last refreshed timestamp
		countries[i].LastRefreshedAt = time.Now().UTC().Format(time.RFC3339)
	}

	// Set global last refreshed timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)
	models.SetLastRefreshedAt(timestamp)

	// Save countries to database
	if err := models.SaveCountries(countries); err != nil {
		return err
	}

	// Generate summary image (only for countries with valid GDP)
	if err := helpers.GenerateSummaryImage(countries, timestamp); err != nil {
		// Log error but don't fail the entire refresh
		fmt.Printf("Warning: Failed to generate summary image: %v\n", err)
	}

	return nil
}

// Handler to get all countries with filtering and sorting
func getAllCountries(context *gin.Context) {
	// Get filter parameters and convert to lowercase
	region := strings.ToLower(context.Query("region"))
	currency := strings.ToUpper(context.Query("currency")) // Currency codes are typically uppercase
	sort := strings.ToLower(context.Query("sort"))

	// Get all countries first
	countries := models.GetAllCountries()

	// Compute GDP for all countries
	for i := range countries {
		countries[i].ComputeEstimatedGDP()
	}

	// Apply filters
	filteredCountries := make([]models.Country, 0)
	for _, country := range countries {
		// Check if country matches all provided filters (case-insensitive)
		matchesRegion := region == "" || strings.ToLower(country.Region) == region
		matchesCurrency := currency == "" || strings.ToUpper(country.CurrencyCode) == currency

		if matchesRegion && matchesCurrency {
			filteredCountries = append(filteredCountries, country)
		}
	}

	// Apply sorting if specified
	if sort != "" {
		switch sort {
		case "gdp_desc":
			// Sort by GDP in descending order
			for i := 0; i < len(filteredCountries)-1; i++ {
				for j := i + 1; j < len(filteredCountries); j++ {
					if filteredCountries[i].EstimatedGDP < filteredCountries[j].EstimatedGDP {
						filteredCountries[i], filteredCountries[j] = filteredCountries[j], filteredCountries[i]
					}
				}
			}
		case "gdp_asc":
			// Sort by GDP in ascending order
			for i := 0; i < len(filteredCountries)-1; i++ {
				for j := i + 1; j < len(filteredCountries); j++ {
					if filteredCountries[i].EstimatedGDP > filteredCountries[j].EstimatedGDP {
						filteredCountries[i], filteredCountries[j] = filteredCountries[j], filteredCountries[i]
					}
				}
			}
		}
	}

	context.JSON(http.StatusOK, filteredCountries)
}

// Handler to get a country by name
func getCountryByName(context *gin.Context) {
	name := context.Param("name") // Get name from URL parameter
	if name == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Country name is required"})
		return
	}

	countries := models.GetAllCountries()

	// Debug: Check if countries are loaded
	if len(countries) == 0 {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "No countries in database. Please call POST /countries/refresh first",
		})
		return
	}

	searchName := strings.ToLower(strings.TrimSpace(name))

	for i := range countries {
		countryName := strings.ToLower(strings.TrimSpace(countries[i].Name))
		if countryName == searchName {
			countries[i].ComputeEstimatedGDP()
			context.JSON(http.StatusOK, countries[i])
			return
		}
	}

	// Debug: Show what we're searching for.
	context.JSON(http.StatusNotFound, gin.H{
		"error":        "Country not found",
		"searched_for": name,
	})
}

// Handler to delete a country by name
func deleteCountryByName(context *gin.Context) {
	name := context.Param("name")
	if name == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Country name is required"})
		return
	}

	err := models.DeleteCountryByName(name)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Country '%s' deleted successfully", name),
	})
}

// Handler to get system status
func getStatus(context *gin.Context) {
	totalCountries := models.GetTotalCountries()
	lastRefreshedAt := models.GetLastRefreshedAt()

	context.JSON(http.StatusOK, gin.H{
		"total_countries":   totalCountries,
		"last_refreshed_at": lastRefreshedAt,
	})
}

// Handler to serve the summary image
func getSummaryImage(context *gin.Context) {
	imagePath := helpers.GetSummaryImagePath()

	// Check if image exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Summary image not found"})
		return
	}

	// Serve the image file
	context.File(imagePath)
}
