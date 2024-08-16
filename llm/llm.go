package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// env variables
var (
	apiKey     = os.Getenv("OPENAI_API_KEY")
	apiBaseURL = os.Getenv("OPENAI_API_BASE_URL")
)

func AnalyzeImage(imgString string, description string) (string, error) {

	prompt := fmt.Sprintf(`Please analyze the uploaded image and provide a score from 0 to 5 indicating how well the image matches the description '%s'. 

Output the results in JSON format with the following structure:

{
	"score": "number between 0 and 5",
	"description": "string explaining how the score was determined"
}

- score: A numerical value between 0 and 5 representing the percentage match between the image and the description.
- description: A textual explanation providing details about how the score was calculated, such as specific features or aspects of the image that influenced the score.`, description)

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiBaseURL
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,

			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: prompt,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: fmt.Sprintf("data:image/jpg;base64,%s", imgString),
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	answer := resp.Choices[0].Message.Content

	return answer, nil
}
