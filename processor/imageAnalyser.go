package processor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/joho/godotenv"
)

func uploadFromPathToGemini(ctx context.Context, client *genai.Client, path, mimeType string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	return uploadFromFileToGemini(ctx, client, file, mimeType)
}

func uploadFromFileToGemini(ctx context.Context, client *genai.Client, file io.Reader, mimeType string) string {
	options := genai.UploadFileOptions{
		DisplayName: "image.png",
		MIMEType:    mimeType,
	}
	fileData, err := client.UploadFile(ctx, "", file, &options)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}

	log.Printf("Uploaded file %s as: %s", fileData.DisplayName, fileData.URI)
	return fileData.URI
}

func getModel(client *genai.Client) *genai.GenerativeModel {
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(1)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"
	return model
}

func Start() string {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()

	apiKey, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		log.Fatalln("Environment variable GEMINI_API_KEY not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	model := getModel(client)

	fileURIs := []string{
		uploadFromPathToGemini(ctx, client, "./images/image1.png", "image/png"),
	}

	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.FileData{URI: fileURIs[0]},
			},
		},
	}

	resp, err := session.SendMessage(ctx, genai.Text("El alias es un string alphanumérico sin espacios. Identifica el alias en esta imagen. Solo devuelve el alias todo en minuscúlas."))
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		fmt.Printf("Alias: %v\n", part)
		return fmt.Sprintf("%v\n", part)
	}

	return "Error processing image"
}

func StartFromFile(file io.Reader, userApiKey string) string {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()
	apiKey := userApiKey
	if len(apiKey) == 0 {
		defaultApi, ok := os.LookupEnv("GEMINI_API_KEY")
		if !ok {
			log.Fatalln("Environment variable GEMINI_API_KEY not set")
		}
		apiKey = defaultApi
		log.Println("Using default API key")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	model := getModel(client)

	fileURIs := []string{
		uploadFromFileToGemini(ctx, client, file, "image/png"),
	}

	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.FileData{URI: fileURIs[0]},
			},
		},
	}

	resp, err := session.SendMessage(ctx, genai.Text("El alias es un string alphanumérico sin espacios ni tildes, no contiene caracteres especiales salvo el punto. Identifica el alias en esta imagen. Solo devuelve el alias todo en minuscúlas sin saltos de linea."))
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		fmt.Printf("Alias: %v\n", part)
		return fmt.Sprintf("%v", part)
	}

	return "Error processing image"
}
