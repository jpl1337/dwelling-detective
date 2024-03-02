package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.POST("/getLocations", func(c *gin.Context) {
		address := c.PostForm("address")

		// Call a function to fetch locations using the Google Maps API
		locations, err := getLocations(address)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, location := range locations {
			fmt.Printf("Name: %s\nVicinity: %s\nLatitude: %f\nLongitude: %f\n\n",
				location.Name, location.Vicinity, location.Geometry.Location.Lat,
				location.Geometry.Location.Lng)
		}
		c.HTML(http.StatusOK, "result.html", gin.H{"locations": locations})
	})

	router.Run(":8080")
}

// Define a struct to hold location details
type Location struct {
	Name       string
	Distance   string
	TravelTime string
}

func getLocations(address string) ([]Location, error) {
	// Convert address to coordinates using Google Maps Geocoding API
	coordinates, err := getCoordinates(address)
	if err != nil {
		return nil, err
	}

	// Call Google Places API to get nearby locations
	apiURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=%f,%f&radius=1500&type=grocery_or_supermarket|airport|bar|restaurant&key=YOUR_GOOGLE_PLACES_API_KEY", coordinates.Lat, coordinates.Lng)
	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var locationsResponse LocationsResponse
	err = json.Unmarshal(body, &locationsResponse)
	if err != nil {
		return nil, err
	}

	return locationsResponse.Results, nil
}

// Function to get coordinates from an address using Google Maps Geocoding API
func getCoordinates(address string) (*Location, error) {
	address = strings.ReplaceAll(address, " ", "+")
	apiURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=YOUR_GOOGLE_MAPS_API_KEY", address)

	response, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var geocodeResponse LocationsResponse
	err = json.Unmarshal(body, &geocodeResponse)
	if err != nil {
		return nil, err
	}

	// Extract the first result (assuming it's the most relevant)
	if len(geocodeResponse.Results) > 0 {
		return &geocodeResponse.Results[0], nil
	}

	return nil, fmt.Errorf("No coordinates found for the given address")
}

// func getLocations(address string) []Location {
// 	// Implement the logic to fetch data from Google Maps API
// 	// Parse the response and return a slice of Location structs
// 	// Include Name, Distance, and TravelTime in each struct
// 	// You'll need to use the Maps JavaScript API for Places and Directions

// 	// Mock data for testing
// 	locations := []Location{
// 		{Name: "Grocery Store 1", Distance: "1.5 miles", TravelTime: "5 minutes"},
// 		{Name: "Airport 1", Distance: "10 miles", TravelTime: "20 minutes"},
// 		{Name: "Bar 1", Distance: "0.8 miles", TravelTime: "3 minutes"},
// 		{Name: "Restaurant 1", Distance: "2.2 miles", TravelTime: "8 minutes"},
// 		{Name: "Grocery Store 2", Distance: "2.8 miles", TravelTime: "10 minutes"},
// 	}

// 	return locations
// }
