package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

// Define a struct to hold location details
type Location struct {
	Name     string `json:"name"`
	Vicinity string `json:"vicinity"`
	Geometry struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
}

// LocationsResponse struct to hold the response from Google Places API
type LocationsResponse struct {
	Results []Location `json:"results"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	router.POST("/getLocations", func(c *gin.Context) {
		address := c.PostForm("address")

		// Call a function to fetch locations using the Google Maps API
		locations, err := getLocations(os.Getenv("GOOGLE_API_KEY"), address)
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

func getLocations(api_key string, address string) ([]Location, error) {
	// Convert address to coordinates using Google Maps Geocoding API
	coordinates, err := getCoordinates(api_key, address)
	if err != nil {
		return nil, err
	}
	c, err := maps.NewClient(maps.WithAPIKey("Insert-API-Key-Here"))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	// Address is the street address that you want to geocode, in the format used by
	// the national postal service of the country concerned.
	// r := &maps.DirectionsRequest{
	// 	Origin:      "Sydney",
	// 	Destination: "Perth",
	// }
	// route, _, err := c.Directions(context.Background(), r)
	// if err != nil {
	// 	log.Fatalf("fatal error: %s", err)
	// }

	// Call Google Places API to get nearby locations
	// apiURL := fmt.Sprintf(
	// 	"https://maps.googleapis.com/maps/api/place/nearbysearch/json?location=",
	// 	"%f,%f&radius=1500&type=grocery_or_supermarket|airport|bar|restaurant&key=%s",
	// 	coordinates.Geometry.Location.Lat, coordinates.Geometry.Location.Lng, api_key)
	// response, err := http.Get(apiURL)
	// fmt.Println("GetLocations response: %s", response)
	// if err != nil {
	// 	return nil, err
	// }
	// defer response.Body.Close()

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
func getCoordinates(c, api_key string, address string) (*Location, error) {
	address = strings.ReplaceAll(address, " ", "+")
	// apiURL := fmt.Sprintf(
	// 	"https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s",
	// 	address, api_key)

	// response, err := http.Get(apiURL)
	// fmt.Println("GetCoordinates response: %s", response)

	// if err != nil {
	// 	return nil, err
	// }
	// defer response.Body.Close()

	// body, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	return nil, err
	// }
	r := &maps.GeocodingRequest{
		Address: address,
	}
	c.Geocode(context.Background(), r)

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

// Test address: 1700, 6 Northside Dr NW Ste A-5,A, Atlanta, GA 30318
