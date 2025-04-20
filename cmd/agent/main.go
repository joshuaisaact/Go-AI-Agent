package main

import (
	"bufio"
	"context"
	"log"
	"os"

	"agent/pkg/agent"
	"agent/pkg/tools"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: ANTHROPIC_API_KEY environment variable not set.")
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	scanner := bufio.NewScanner(os.Stdin)

	var getUserMessage agent.MessageHandler = func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agentInstance := agent.NewAgent(&client, getUserMessage, tools.GetTools())
	err := agentInstance.Run(context.TODO())
	if err != nil {
		log.Printf("Agent exited with error: %s\n", err.Error())
	}
}