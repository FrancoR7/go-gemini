package processor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"errors"

	"github.com/joho/godotenv"
)

func uploadFromFileToGemini(ctx context.Context, client *genai.Client, file io.Reader, mimeType string) (string, error) {
	options := genai.UploadFileOptions{
		DisplayName: "image.png",
		MIMEType:    mimeType,
	}
	fileData, err := client.UploadFile(ctx, "", file, &options)
	if err != nil {
		log.Printf("Error uploading file: %v \n", err)
		return "", fmt.Errorf("error uploading file: %v", err)
	}

	fmt.Printf("Uploaded file %s as: %s \n", fileData.DisplayName, fileData.URI)
	return fileData.URI, nil
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

func StartFromFile(file io.Reader, userApiKey string) (string, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	ctx := context.Background()
	apiKey := userApiKey
	if len(apiKey) == 0 {
		defaultApi, ok := os.LookupEnv("GEMINI_API_KEY")
		if !ok {
			log.Fatalln("Environment variable GEMINI_API_KEY not set")
		}
		apiKey = defaultApi
		fmt.Println("Using default API key")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer client.Close()

	model := getModel(client)

	fileURI, error := uploadFromFileToGemini(ctx, client, file, "image/png")
	if error != nil {
		return "", error
	}
	fileURIs := []string{
		fileURI,
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

	resp, errorMessage := session.SendMessage(ctx, genai.Text("El `alias` es un string alfanumérico donde no se permiten espacios ni caracteres especiales, excepto el punto, los espacios al principio, al final o entre las palabras no estan permitidos. Identifica un alias en la imagen provista y extraelo retornando unicamente el string identificado, transformando todas las letras a minuscúlas."))
	if errorMessage != nil {
		fmt.Printf("Error sending message: %v \n", errorMessage)
		return "", fmt.Errorf("error sending message: %v", errorMessage)
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		fmt.Printf("Alias: %v\n", part)
		return strings.Replace(fmt.Sprintf("%v", part), "\n", "", 1), nil
	}

	return "", errors.New("error processing image")
}
