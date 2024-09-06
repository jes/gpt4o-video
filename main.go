package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory_path>")
		return
	}

	dirPath := os.Args[1]
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)

	images, err := loadImagesFromDirectory(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	if len(images) == 0 {
		log.Fatal("No PNG images found in the specified directory")
	}

	fmt.Printf("Loaded %d PNG images\n", len(images))

	err = interactiveQA(images, client)
	if err != nil {
		log.Fatal(err)
	}
}

func loadImagesFromDirectory(dirPath string) ([]openai.ChatMessagePart, error) {
	var images []openai.ChatMessagePart

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.ToLower(filepath.Ext(file.Name())) == ".png" {
			imagePath := filepath.Join(dirPath, file.Name())
			imageData, err := ioutil.ReadFile(imagePath)
			if err != nil {
				return nil, fmt.Errorf("error reading image file %s: %v", imagePath, err)
			}

			images = append(images, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData),
				},
			})
		}
	}

	return images, nil
}

func interactiveQA(images []openai.ChatMessagePart, client *openai.Client) error {
	reader := bufio.NewReader(os.Stdin)

	// Initialize the dialogue with the images
	dialogue := []openai.ChatCompletionMessage{
		{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: images,
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "The images you are seeing represent frames from a video.",
		},
	}

	for {
		fmt.Print("> ")
		question, _ := reader.ReadString('\n')
		question = strings.TrimSpace(question)

		if strings.ToLower(question) == "quit" {
			break
		}

		// Append the user's question to the dialogue
		dialogue = append(dialogue, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: question,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:     "gpt-4o",
				Messages:  dialogue,
				MaxTokens: 300,
			},
		)

		if err != nil {
			return fmt.Errorf("error calling GPT-4 Vision API: %v", err)
		}

		// Append the assistant's response to the dialogue
		dialogue = append(dialogue, resp.Choices[0].Message)

		fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
	}

	return nil
}
