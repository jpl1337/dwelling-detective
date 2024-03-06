package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

// struct to hold coordinates
type Coordinate struct {
	Latitude  float64 `json:"latitude`
	Longitude float64 `json:"longitude`
}

// struct to hold location details
type Location struct {
	Name        string     `json:"name"`
	Vicinity    string     `json:"vicinity"`
	Coordinates Coordinate `json:"coordinates"`
}

// struct to hold the response from Google Places API
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
	// TODO: make this function return a Coordinates struct
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

func getNearby(target Coordinate) {
	// TODO: this needs to grab nearby locations that we're looking for
	// the google nearby api is limited to 50km or about 31miles
	// if airports are not found in this radius..
	// find nearest airport from csv
	findNearestAirport(target)
}

func findNearestAirport(target Coordinate) {
	filePath := filepath.Join("data", "airports.csv")

	allCoordinates, err := loadCoordinates(filePath)
	if err != nil {
		log.Fatal(err)
	}
	closestConcurrent := findClosestCoordinates(target, allCoordinates)
	fmt.Printf("Concurrent: Closest Coordinate: %v\n", closestConcurrent)
}

func findClosestCoordinates(target Coordinate, allCoordinates []Coordinate) Coordinate {
	numCPU := runtime.NumCPU()
	chunkSize := (len(allCoordinates) + numCPU - 1) / numCPU // Divide coordinates into chunks for concurrent processing

	var closestCoordinate Coordinate
	closestDistance := math.MaxFloat64

	var wg sync.WaitGroup
	wg.Add(numCPU)

	for i := 0; i < numCPU; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > len(allCoordinates) {
			end = len(allCoordinates)
		}

		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				distance := haversineDistance(target.Latitude, target.Longitude, allCoordinates[j].Latitude, allCoordinates[j].Longitude)
				if distance < closestDistance {
					closestCoordinate = allCoordinates[j]
					closestDistance = distance
				}
			}
		}(start, end)
	}

	wg.Wait()

	return closestCoordinate
}

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Earth radius in kilometers

	// Convert latitude and longitude from degrees to radians
	lat1, lon1, lat2, lon2 = deg2rad(lat1), deg2rad(lon1), deg2rad(lat2), deg2rad(lon2)

	// Haversine formula
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	distance := earthRadius * c

	return distance
}

func parseCoordinate(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

func loadCoordinates(filename string) ([]Coordinate, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var coordinates []Coordinate
	for _, record := range records {
		latitude, _ := parseCoordinate(record[0])
		longitude, _ := parseCoordinate(record[1])
		coordinates = append(coordinates, Coordinate{Latitude: latitude, Longitude: longitude})
	}

	return coordinates, nil
}

// Test address: 1700, 6 Northside Dr NW Ste A-5,A, Atlanta, GA 30318
