package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jhw/go-outrights/pkg/outrights"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request: %s", request.Body)
	
	var req outrights.SimulationRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("Error unmarshaling request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid JSON",
		}, nil
	}
	
	result := outrights.ProcessSimulation(req)
	
	responseBody, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func runCLI() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run . <filename>")
	}
	
	filename := os.Args[1]
	
	// Read and parse the JSON file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	
	var events []outrights.Event
	if err := json.Unmarshal(data, &events); err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Processing %s with %d events", filename, len(events))
	log.Println("Starting simulation...")
	
	result := outrights.ProcessEventsFile(events)
	
	log.Printf("Home advantage: %.4f, Solver error: %.6f", result.HomeAdvantage, result.SolverError)
	log.Println()
	log.Println("Teams (sorted by points per game rating):")
	for _, team := range result.Teams {
		log.Printf("- %s: %.1f pts, PPG rating: %.3f", team.Name, team.Points, team.PointsPerGameRating)
	}
	
	log.Println()
	log.Println("Outright marks:")
	for _, mark := range result.OutrightMarks {
		log.Printf("- %s: %.3f", mark.Team, mark.Mark)
	}
}

func main() {
	// Check if running locally (not in Lambda)
	if len(os.Args) > 1 {
		runCLI()
		return
	}
	
	lambda.Start(handleRequest)
}