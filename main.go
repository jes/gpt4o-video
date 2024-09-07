package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <video_path> <fps>\nfps is optional, default is 1")
		return
	}

	videoPath := os.Args[1]
	fps := 1.0 // Default FPS
	if len(os.Args) > 2 {
		var err error
		fps, err = strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			log.Fatal("Invalid FPS value")
		}
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(apiKey)

	fmt.Printf("Splitting video into frames at %f fps...\n", fps)
	framesDir, err := splitVideoIntoFrames(videoPath, fps)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(framesDir) // Ensure the temporary directory is deleted

	fmt.Println("Loading images from directory...")
	messages, err := loadImagesFromDirectory(framesDir, fps)
	if err != nil {
		log.Fatal(err)
	}

	if len(messages) == 0 {
		log.Fatal("No PNG images found in the specified directory")
	}

	fmt.Printf("Loaded %d PNG images\n", len(messages)/2)

	err = interactiveQA(messages, client)
	if err != nil {
		log.Fatal(err)
	}

	// Unlink the temporary frames directory
	err = os.RemoveAll(framesDir)
	if err != nil {
		log.Printf("Warning: Failed to remove temporary directory %s: %v", framesDir, err)
	}
}

func splitVideoIntoFrames(videoPath string, fps float64) (string, error) {
	framesDir, err := os.MkdirTemp("", "frames")
	if err != nil {
		return "", fmt.Errorf("error creating temporary frames directory: %v", err)
	}

	fpsString := fmt.Sprintf("%f", fps)
	outputSize := "512:512"
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", fmt.Sprintf("fps=%s,scale=%s:force_original_aspect_ratio=decrease,pad=%s:(ow-iw)/2:(oh-ih)/2", fpsString, outputSize, outputSize), filepath.Join(framesDir, "frame_%04d.png"))
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error splitting video into frames: %v", err)
	}

	return framesDir, nil
}

func loadImagesFromDirectory(dirPath string, fps float64) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.ToLower(filepath.Ext(file.Name())) == ".png" {
			imagePath := filepath.Join(dirPath, file.Name())
			imageData, err := os.ReadFile(imagePath)
			if err != nil {
				return nil, fmt.Errorf("error reading image file %s: %v", imagePath, err)
			}

			timestamp := extractTimestampFromFilename(file.Name(), fps)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: "",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData),
						},
					},
				},
			}, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: timestamp,
			})
		}
	}

	return messages, nil
}

func extractTimestampFromFilename(filename string, fps float64) string {
	// Assuming filename format is frame_XXXX.png
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return ""
	}
	frameNumber := strings.TrimSuffix(parts[1], ".png")
	frameNum, err := strconv.Atoi(frameNumber)
	if err != nil {
		return ""
	}
	seconds := float64(frameNum) / fps
	return fmt.Sprintf("Timestamp: %02d:%02d", int(seconds)/60, int(seconds)%60)
}

func interactiveQA(messages []openai.ChatCompletionMessage, client *openai.Client) error {
	reader := bufio.NewReader(os.Stdin)

	// Initialize the dialogue with the images and timestamps
	dialogue := append([]openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "The images you are seeing represent frames from a video.",
		},
	}, messages...)

	for {
		fmt.Print("> ")
		question, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading input: %v", err)
		}
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
				MaxTokens: 4096,
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
